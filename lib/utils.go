package lib

import (
	"os/exec"
	"fmt"
	"strings"
)

func printCmd(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func RunCmd(cmd *exec.Cmd) (string, bool) {
	printCmd(cmd)

	resultBytes, _ := cmd.CombinedOutput()
	result := string(resultBytes[:])

	return result, cmd.ProcessState.Success()
}
