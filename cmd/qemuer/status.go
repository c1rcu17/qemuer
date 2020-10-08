package main

import (
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

var statusTemplate = template.Must(template.New("").Parse(strings.TrimLeft(`
File:      {{ .File }}
Home:      {{ .Home }}
Name:      {{ .Name }}
Arch:      {{ .Arch }}
Bios:      {{ .Bios }}{{ if eq .Bios "uefi" }} ({{ .BiosFile }}){{ end }}
CPU:       {{ .CPU.Sockets }}-{{ .CPU.Cores }}-{{ .CPU.Threads }}
Memory:    {{ .Memory }} Mb
ISO:       {{ if .ISO }}{{ .ISO }}{{ else }}-{{ end }}
Disks:     {{ range $i, $d := .Disks }}
{{- if ne $i 0 }}           {{ end }}{{ $d }}
{{ end -}}
Networks:  {{ range $i, $n := .Networks }}
{{- if ne $i 0 }}
	   {{ end }}Name:      {{ $n.Name }}
	   BridgeDev: {{ $n.BridgeDev }}
	   NatDev:    {{ if $n.NatDev }}{{ $n.NatDev }}{{ else }}-{{ end }}
           MAC:       {{ $n.MAC }}
           Subnet:    {{ $n.Subnet }}
           Netmask:   {{ $n.Netmask }}
	   Gateway:   {{ $n.Gateway }}
	   Broadcast: {{ $n.Broadcast }}
	   IP Range:  {{ $n.IPStart }} - {{ $n.IPEnd }}
{{ end -}}
Video:     {{ if ne .Video "none" }}{{ .Video }}{{ else }}-{{ end }}{{ if eq .Video "qxl" }} ({{ .Display }}){{ end }}
Monitor:   {{ .Monitor }}
Console:   {{ .Console }}
PIDFile:   {{ .PIDFile }}
PID:       {{ .PID }}
`, "\n")))

func statusCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	if err := statusTemplate.Execute(os.Stdout, ec); err != nil {
		return err
	}

	return nil
}
