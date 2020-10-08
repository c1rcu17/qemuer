package main

import (
	"github.com/urfave/cli/v2"
)

func killCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	if err := monitorCommand(ec, "quit"); err != nil {
		return err
	}

	return nil
}
