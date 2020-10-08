package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"

	"github.com/c1rcu17/qemuer/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func prepareConfig(ctx *cli.Context) (*config.EnrichedConfig, error) {
	yamlFile := ctx.String("file")
	yamlData, err := ioutil.ReadFile(yamlFile)

	if err != nil {
		return nil, err
	}

	c := config.NewConfig()

	if err = yaml.UnmarshalStrict(yamlData, c); err != nil {
		return nil, err
	}

	ec, err := config.NewEnrichedConfig(c, yamlFile)

	if err != nil {
		return nil, err
	}

	return ec, nil
}

func execv(ctx *cli.Context, prog config.Prog, args []string) error {
	args = append([]string{prog.Name}, args...)

	if ctx.Bool("dry-run") {
		fmt.Println(strings.Join(args, " "))
	} else {
		if err := syscall.Exec(prog.Path, args, syscall.Environ()); err != nil {
			return err
		}
	}

	return nil
}

func monitorCommand(ec *config.EnrichedConfig, cmd string) error {
	stdin := &bytes.Buffer{}

	if _, err := stdin.WriteString(cmd + "\n"); err != nil {
		return err
	}

	socat := exec.Command(ec.Progs.Socat.Path, "-", fmt.Sprintf("UNIX-CONNECT:%s", ec.Monitor))
	socat.Stdin = stdin

	if err := socat.Run(); err != nil {
		return err
	}

	return nil
}
