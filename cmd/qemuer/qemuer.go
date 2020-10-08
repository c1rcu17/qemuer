package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	VMFlags := []cli.Flag{
		&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: "print commands instead of executing"},
		&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Required: true, Usage: "name of the `VMFILE`"},
	}

	app := &cli.App{
		Name:  "qemuer",
		Usage: "launch QEMU virtual machines like if you know how to do it",
		Commands: []*cli.Command{
			{Name: "console", Aliases: []string{"c"}, Flags: VMFlags, Action: consoleCmd, Usage: "Connect to the virtual machine' serial console"},
			{Name: "display", Aliases: []string{"d"}, Flags: VMFlags, Action: displayCmd, Usage: "Connect to the virtual machine's QXL display"},
			{Name: "kill", Aliases: []string{"k"}, Flags: VMFlags, Action: killCmd, Usage: "Force shutdown the virtual machine"},
			{Name: "monitor", Aliases: []string{"m"}, Flags: VMFlags, Action: monitorCmd, Usage: "Connect to the virtual machine's QEMU monitor"},
			{Name: "poweroff", Aliases: []string{"p"}, Flags: VMFlags, Action: poweroffCmd, Usage: "Gracefully shutdown the virtual machine"},
			{Name: "run", Aliases: []string{"r"}, Flags: VMFlags, Action: runCmd, Usage: "Turn on the virtual machine"},
			{Name: "status", Aliases: []string{"s"}, Flags: VMFlags, Action: statusCmd, Usage: "Print the status of the virtual machine"},
			{Name: "version", Aliases: []string{"v"}, Action: versionCmd, Usage: "Print the version and exit"},
		},
		ExitErrHandler: func(ctx *cli.Context, err error) {},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "\033[31mError:\033[0m", err.Error())
		os.Exit(1)
	}
}
