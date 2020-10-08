package util

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func GetFDs() int {
	pid := strconv.Itoa(os.Getpid())
	cmd := exec.Command("/usr/bin/lsof", "-p", pid)

	if out, err := cmd.Output(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot execute lsof for pid %s\n%v\n", pid, err)
		return 0
	} else {
		lines := strings.Split(string(out), "\n")
		return len(lines) - 1
	}
}

func PrintFDLeaks(fds int) {
	fmt.Println("Number of FD leaks:", GetFDs()-fds)
}
