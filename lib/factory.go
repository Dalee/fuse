package lib

import (
	"fmt"
	"os/exec"
)

func CommandFactory(context string, args[] string) *exec.Cmd {
	argList := make([]string, 0)
	if context != "" {
		argList = append(argList, fmt.Sprintf("--context=%s", context))
	}

	argList = append(argList, args...)
	return exec.Command("kubectl", argList...)
}
