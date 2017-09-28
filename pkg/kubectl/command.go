package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type (
	kubeCommandInterface interface {
		Run() ([]byte, bool)
		getCommand() *exec.Cmd
	}

	kubeCommand struct {
		cmd *exec.Cmd
	}
)

// Easy to use wrapper
func newCommand(args []string) kubeCommandInterface {
	return newCommandWithBinary(args, "kubectl")
}

// More advanced wrapper which allows you to override binary to run
func newCommandWithBinary(args []string, binary string) *kubeCommand {
	argList := make([]string, 0)
	if context := os.Getenv(ClusterContextEnv); context != "" {
		argList = append(argList, fmt.Sprintf("--context=%s", context))
	}

	argList = append(argList, args...)
	return &kubeCommand{
		cmd: exec.Command(binary, argList...),
	}
}

// Execute command and get stdout, stderr and exit_code as bool
func (c *kubeCommand) Run() ([]byte, bool) {
	fmt.Printf("===> %s\n", strings.Join(c.getCommand().Args, " ")) // TODO: should be moved to logging

	result, err := c.getCommand().CombinedOutput()
	if len(result) == 0 {
		result = []byte("Command failed to run")
	}

	return result, err == nil
}

// Command getter
func (c *kubeCommand) getCommand() *exec.Cmd {
	return c.cmd
}
