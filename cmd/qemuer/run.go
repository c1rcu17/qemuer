package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/c1rcu17/qemuer/config"
	"github.com/c1rcu17/qemuer/static"
	"github.com/urfave/cli/v2"
)

type netDumpXML struct {
	Forward struct {
		Dev string `xml:"dev,attr"`
	} `xml:"forward"`
	Bridge struct {
		Name string `xml:"name,attr"`
	} `xml:"bridge"`
	IP struct {
		Address string `xml:"address,attr"`
		Netmask string `xml:"netmask,attr"`
		DHCP    struct {
			Range struct {
				Start string `xml:"start,attr"`
				End   string `xml:"end,attr"`
			} `xml:"range"`
		} `xml:"dhcp"`
	} `xml:"ip"`
}

var networkTemplate = template.Must(template.New("").Parse(strings.TrimLeft(`
<network>
    <name>{{ .Name }}</name>
    <forward mode='nat'{{ if .NatDev }} dev='{{ .NatDev }}'{{ end }}/>
    <bridge name='{{ .BridgeDev }}'/>
    <ip address='{{ .Gateway }}' netmask='{{ .Netmask }}'>
        <dhcp>
            <range start='{{ .IPStart }}' end='{{ .IPEnd }}'/>
        </dhcp>
    </ip>
</network>
`, "\n")))

func runCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(ec.Runtime, 0755); err != nil {
		return err
	}

	bootIndex := 0
	bootOrder := ""
	bootMenu := "off"

	qemuArgs := []string{
		"-name", ec.Name,
		"-nodefaults", "-no-user-config", "-no-hpet",
		"-machine", "q35,accel=kvm,vmport=off,dump-guest-core=off"}

	if ec.Bios == config.BiosUEFI {
		if _, err := os.Stat(ec.BiosFile); err != nil {
			if os.IsNotExist(err) {
				if err := installBios(ec.BiosFile); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		qemuArgs = append(qemuArgs, "-bios", ec.BiosFile)
	}

	qemuArgs = append(qemuArgs, "-cpu", "host",
		"-smp", fmt.Sprintf("%d,sockets=%d,cores=%d,threads=%d",
			ec.CPU.Sockets*ec.CPU.Cores*ec.CPU.Threads,
			ec.CPU.Sockets, ec.CPU.Cores, ec.CPU.Threads),
		"-m", strconv.Itoa(ec.Memory),
		"-chardev", fmt.Sprintf("socket,id=char0,path=%s,server,nowait", ec.Console),
		"-device", "isa-serial,chardev=char0",
		"-chardev", fmt.Sprintf("socket,id=char1,path=%s,server,nowait", ec.Monitor),
		"-mon", "chardev=char1",
		"-object", "rng-random,id=obj0,filename=/dev/urandom",
		"-device", "virtio-rng-pci,rng=obj0",
		"-device", "virtio-balloon-pci",
		"-pidfile", ec.PIDFile,
		"-daemonize",
		"-k", "pt",
	)

	if len(ec.ISOs) > 0 {
		for i, f := range ec.ISOs {
			qemuArgs = append(qemuArgs,
				"-drive", fmt.Sprintf("id=drive%d,if=none,format=raw,file=%s", i, f),
				"-device", fmt.Sprintf("ide-cd,drive=drive%d,bus=ide.%d,bootindex=%d", i, i+1, bootIndex))
			bootIndex++
		}

		bootOrder = bootOrder + "c"
	}

	if len(ec.Disks) > 0 {
		for i, d := range ec.Disks {
			qemuArgs = append(qemuArgs,
				"-blockdev", fmt.Sprintf("qcow2,node-name=block%d,file.driver=file,file.filename=%s", i, d),
				"-device", fmt.Sprintf("virtio-blk-pci,drive=block%d,bootindex=%d", i, bootIndex))
			bootIndex++
		}

		bootOrder = bootOrder + "d"
	}

	for i, n := range ec.Networks {
		if err := createNetwork(&n, ec.Progs.Virsh); err != nil {
			return err
		}

		qemuArgs = append(qemuArgs,
			"-netdev", fmt.Sprintf("bridge,id=net%d,br=%s", i, n.BridgeDev),
			"-device", fmt.Sprintf("virtio-net-pci,netdev=net%d,mac=%s", i, n.MAC))
	}

	if ec.Video == config.VideoNone {
		qemuArgs = append(qemuArgs, "-nographic")
	} else {
		qemuArgs = append(qemuArgs,
			"-device", "ich9-usb-ehci1,id=usb",
			"-device", "ich9-usb-uhci1,masterbus=usb.0,firstport=0,multifunction=on",
			"-device", "ich9-usb-uhci2,masterbus=usb.0,firstport=2",
			"-device", "ich9-usb-uhci3,masterbus=usb.0,firstport=4",
			"-device", "usb-tablet")

		switch ec.Video {
		case config.VideoQXL:
			qemuArgs = append(qemuArgs,
				"-device", "qxl-vga,vgamem_mb=64,max_outputs=1",
				"-spice", fmt.Sprintf("addr=%s,unix,disable-ticketing,image-compression=off,seamless-migration=on", ec.Display),
				"-chardev", "spicevmc,id=char2,debug=0,name=vdagent",
				"-device", "virtio-serial-pci",
				"-device", "virtserialport,chardev=char2,name=com.redhat.spice.0",
				"-chardev", "spicevmc,id=char3,debug=0,name=usbredir",
				"-device", "usb-redir,chardev=char3",
				"-chardev", "spicevmc,id=char4,debug=0,name=usbredir",
				"-device", "usb-redir,chardev=char4",
				"-chardev", "spicevmc,id=char5,debug=0,name=usbredir",
				"-device", "usb-redir,chardev=char5")
		case config.VideoVGA:
			qemuArgs = append(qemuArgs, "-device", "VGA,vgamem_mb=64")
		case config.VideoVirtIO:
			qemuArgs = append(qemuArgs, "-device", "virtio-gpu-pci")
		}
	}

	switch {
	case bootIndex > 1:
		bootMenu = "on"
		fallthrough
	case bootIndex > 0:
		qemuArgs = append(qemuArgs, "-boot", fmt.Sprintf("order=%s,menu=%s", bootOrder, bootMenu))
	}

	if err := execv(ctx, ec.Progs.Qemu, qemuArgs); err != nil {
		return err
	}

	return nil
}

func createNetwork(en *config.EnrichedNetwork, virsh config.Prog) error {
	if out, err := exec.Command(virsh.Path, "net-dumpxml", en.Name).CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "Network not found") {
			return err
		}
	} else {
		var net netDumpXML

		if err := xml.Unmarshal(out, &net); err != nil {
			return err
		}

		if net.Forward.Dev == en.NatDev &&
			net.Bridge.Name == en.BridgeDev &&
			net.IP.Address == en.Gateway &&
			net.IP.Netmask == en.Netmask &&
			net.IP.DHCP.Range.Start == en.IPStart &&
			net.IP.DHCP.Range.End == en.IPEnd {
			return nil
		}

		return fmt.Errorf("network %s exists with different settings", en.Name)
	}

	f, err := ioutil.TempFile("", "net-*.xml")

	if err != nil {
		return err
	}

	defer os.Remove(f.Name())

	if err := networkTemplate.Execute(f, en); err != nil {
		f.Close()
		return err
	}

	f.Close()

	if err := exec.Command(virsh.Path, "net-create", f.Name()).Run(); err != nil {
		return err
	}

	return nil
}

func installBios(path string) error {
	resource := "/OVMF-pure-efi.fd"

	if src, err := static.OpenResource(resource); err != nil {
		return err
	} else {
		if dst, err := os.Create(path); err != nil {
			return err
		} else {
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return err
			}
		}
	}

	return nil
}
