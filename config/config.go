package config

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/c1rcu17/qemuer/util"
)

type (
	Config struct {
		Name     string
		Arch     Arch
		Bios     Bios
		CPU      CPU
		Memory   int
		ISO      string
		Disks    []string
		Networks []Network
		Video    Video
	}

	Arch string
	Bios string

	CPU struct {
		Sockets, Cores, Threads int
	}

	Network struct {
		NatDev, MAC, CIDR string
	}

	Video string

	EnrichedConfig struct {
		Config
		Networks []EnrichedNetwork
		File     string
		Home     string
		Runtime  string
		Monitor  string
		Console  string
		Display  string
		PIDFile  string
		BiosFile string
		PID      int
		Progs    Progs
	}

	EnrichedNetwork struct {
		Network
		Name, BridgeDev, Subnet, Netmask, Gateway, Broadcast, IPStart, IPEnd string
	}

	Progs struct {
		Qemu    Prog
		Virsh   Prog
		Minicom Prog
		Spicy   Prog
		Socat   Prog
	}

	Prog struct {
		Name string
		Path string
	}
)

const (
	ArchX8664   Arch  = "x86_64"
	BiosLegacy  Bios  = "legacy"
	BiosUEFI    Bios  = "uefi"
	VideoNone   Video = "none"
	VideoQXL    Video = "qxl"
	VideoVGA    Video = "vga"
	VideoVirtIO Video = "virtio"
)

func (p *Prog) Which() error {
	if path, err := exec.LookPath(p.Name); err != nil {
		return err
	} else {
		p.Path = path
	}

	return nil
}

func NewConfig() *Config {
	return &Config{
		Arch:   ArchX8664,
		Bios:   BiosUEFI,
		CPU:    CPU{Sockets: 1, Cores: 2, Threads: 1},
		Memory: 1024,
		Video:  VideoNone,
	}
}

func NewEnrichedConfig(c *Config, file string) (*EnrichedConfig, error) {
	ec := &EnrichedConfig{Config: *c}

	if abs, err := filepath.Abs(file); err != nil {
		return nil, err
	} else {
		ec.File = abs
	}

	ec.Home = filepath.Dir(ec.File)

	if len(ec.Name) < 1 {
		return nil, fmt.Errorf("name field cannot be empty")
	}

	switch ec.Arch {
	case ArchX8664:
		ec.Progs.Qemu.Name = "qemu-system-x86_64"
	default:
		return nil, fmt.Errorf("invalid arch %s, choose from: %v", ec.Arch, []Arch{ArchX8664})
	}

	switch ec.Bios {
	case BiosLegacy, BiosUEFI:
	default:
		return nil, fmt.Errorf("invalid bios %s, choose from: %v", ec.Bios, []Bios{BiosLegacy, BiosUEFI})

	}

	if ec.CPU.Sockets < 1 {
		return nil, fmt.Errorf("cpu.sockets must be greater than 1")
	}

	if ec.CPU.Cores < 1 {
		return nil, fmt.Errorf("cpu.cores must be greater than 1")
	}

	if ec.CPU.Threads < 1 {
		return nil, fmt.Errorf("cpu.threads must be greater than 1")
	}

	if ec.Memory < 64 {
		return nil, fmt.Errorf("memory must be greater than 64")
	}

	if len(ec.ISO) > 0 {
		if !filepath.IsAbs(ec.ISO) {
			ec.ISO = path.Join(ec.Home, ec.ISO)
		}

		if _, err := os.Stat(ec.ISO); err != nil {
			return nil, err
		}
	}

	for i, d := range ec.Disks {
		if !filepath.IsAbs(d) {
			d = path.Join(ec.Home, d)
			ec.Disks[i] = d
		}

		if _, err := os.Stat(d); err != nil {
			return nil, err
		}
	}

	if err := enrichNetworks(ec); err != nil {
		return nil, err
	}

	switch ec.Video {
	case VideoNone, VideoQXL, VideoVGA, VideoVirtIO:
	default:
		return nil, fmt.Errorf("invalid video %s, choose from: %v", ec.Video, []Video{VideoNone, VideoQXL, VideoVGA, VideoVirtIO})
	}

	id := fmt.Sprintf("%x", sha256.Sum256([]byte(ec.File)))[:8]
	ec.Runtime = path.Join("/var/run/qemuer", id)
	ec.Monitor = path.Join(ec.Runtime, "monitor.sock")
	ec.Console = path.Join(ec.Runtime, "console.sock")
	ec.Display = path.Join(ec.Runtime, "display.sock")
	ec.BiosFile = path.Join(ec.Runtime, "bios.bin")
	ec.PIDFile = path.Join(ec.Runtime, "qemu.pid")

	if pid, err := ioutil.ReadFile(ec.PIDFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		ec.PID, err = strconv.Atoi(strings.TrimSpace(string(pid)))

		if err != nil {
			return nil, err
		}
	}

	ec.Progs.Virsh.Name = "virsh"
	ec.Progs.Minicom.Name = "minicom"
	ec.Progs.Spicy.Name = "spicy"
	ec.Progs.Socat.Name = "socat"

	for _, p := range []*Prog{&ec.Progs.Qemu, &ec.Progs.Virsh, &ec.Progs.Minicom, &ec.Progs.Spicy, &ec.Progs.Socat} {
		if err := p.Which(); err != nil {
			return nil, err
		}
	}

	return ec, nil
}

func enrichNetworks(ec *EnrichedConfig) error {
	var interfaces []string
	var macs []string

	if ifaces, err := net.Interfaces(); err != nil {
		return err
	} else {
		for _, i := range ifaces {
			interfaces = append(interfaces, i.Name)

			if i.HardwareAddr != nil {
				macs = append(macs, i.HardwareAddr.String())
			}
		}
	}

	for _, n := range ec.Config.Networks {
		en := EnrichedNetwork{Network: n}

		if mac, err := net.ParseMAC(en.MAC); err != nil {
			return err
		} else {
			if len(en.NatDev) > 0 {
				for i, iface := range interfaces {
					if en.NatDev == iface {
						break
					}
					if i == len(interfaces)-1 {
						return fmt.Errorf("invalid natdev %s, choose from: %v", en.NatDev, interfaces)
					}
				}
			}

			en.MAC = mac.String()

			for _, m := range macs {
				if en.MAC == m {
					return fmt.Errorf("address %s: already in use", en.MAC)
				}
			}

			macs = append(macs, en.MAC)

			first_byte := mac[0]

			if first_byte&0b01 != 0 {
				return fmt.Errorf("address %s: is a multicast MAC address. see: "+
					"https://en.wikipedia.org/wiki/MAC_address#Unicast_vs._multicast", en.MAC)
			}

			if first_byte&0b10 == 0 {
				return fmt.Errorf("address %s: is a universally administered MAC address (UAA). see: "+
					"https://en.wikipedia.org/wiki/MAC_address#Universal_vs._local", en.MAC)
			}
		}

		if addr, subnet, err := net.ParseCIDR(en.CIDR); err != nil {
			return err
		} else {
			en.Subnet = subnet.IP.String()
			en.Netmask = net.IP(subnet.Mask).String()

			if !addr.Equal(subnet.IP) {
				return fmt.Errorf("invalid subnet address %s: it should be %s", addr, en.Subnet)
			}

			if gw, bc, start, end, err := util.AddressRange(subnet); err != nil {
				return err
			} else {
				en.Gateway = gw.String()
				en.Broadcast = bc.String()
				en.IPStart = start.String()
				en.IPEnd = end.String()
			}
		}

		id := fmt.Sprintf("%x", sha256.Sum256([]byte(en.Subnet+en.Netmask)))[:8]
		en.Name = fmt.Sprintf("net-%s", id)
		en.BridgeDev = fmt.Sprintf("br-%s", id)

		ec.Networks = append(ec.Networks, en)
	}

	return nil
}
