package main

import (
	"github.com/urfave/cli/v2"
)

func poweroffCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	if err := monitorCommand(ec, "system_powerdown"); err != nil {
		return err
	}

	return nil
}
