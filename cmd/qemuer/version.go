//go:generate ./versiongen.sh

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

func versionCmd(ctx *cli.Context) error {
	fmt.Println(path.Base(os.Args[0]), versionNumber)
	return nil
}
