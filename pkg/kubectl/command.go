package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type (
	kubeCommandInterface interface {
		Log()
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
	if context := os.Getenv("CLUSTER_CONTEXT"); context != "" {
		argList = append(argList, fmt.Sprintf("--context=%s", context))
	}

	argList = append(argList, args...)
	return &kubeCommand{
		cmd: exec.Command(binary, argList...),
	}
}

// Log command to stdout
func (c *kubeCommand) Log() {
	fmt.Printf("==> Executing: %s\n", strings.Join(c.getCommand().Args, " "))
}

// Execute command and get stdout, stderr and exit_code as bool
func (c *kubeCommand) Run() ([]byte, bool) {
	result, err := c.getCommand().CombinedOutput()
	if err != nil {
		return []byte("Command failed to run"), false
	}

	return result, c.cmd.ProcessState.Success()
}

// Command getter
func (c *kubeCommand) getCommand() *exec.Cmd {
	return c.cmd
}
