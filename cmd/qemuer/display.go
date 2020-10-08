package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func displayCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	spicyArgs := []string{
		fmt.Sprintf("--uri=spice+unix://%s", ec.Display),
		fmt.Sprintf("--title=%s", ec.Name)}

	fmt.Println("Shift+F12 - exit fullscreen")

	if err := execv(ctx, ec.Progs.Spicy, spicyArgs); err != nil {
		return err
	}

	return nil
}
