package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func monitorCmd(ctx *cli.Context) error {
	ec, err := prepareConfig(ctx)

	if err != nil {
		return err
	}

	minicomArgs := []string{"-D", fmt.Sprintf("unix#%s", ec.Monitor)}

	if err := execv(ctx, ec.Progs.Minicom, minicomArgs); err != nil {
		return err
	}

	return nil
}
