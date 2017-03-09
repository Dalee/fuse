package kubectl

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

// ensure command will be created without cluster context
func TestCreateCommand(t *testing.T) {
	cliCommand := newCommand([]string{
		"hello",
		"world",
	})

	cmdString := strings.Join(cliCommand.getCommand().Args, " ")
	assert.Equal(t, "kubectl hello world", cmdString)
}

// ensure command will be create with cluster context
func TestCreateCommandContext(t *testing.T) {
	os.Setenv(ClusterContextEnv, "live-context")
	cliCommand := newCommand([]string{
		"hello",
		"world",
	})

	cmdString := strings.Join(cliCommand.getCommand().Args, " ")
	assert.Equal(t, "kubectl --context=live-context hello world", cmdString)
	os.Unsetenv(ClusterContextEnv)
}

// ensure command can be executed
func TestExecuteCommand(t *testing.T) {
	cliCommand := newCommandWithBinary([]string{"/"}, "ls")
	_, ok := cliCommand.Run()
	assert.True(t, ok)
}

func TestExecuteCommandFailed(t *testing.T) {
	cliCommand := newCommandWithBinary([]string{"/"}, "_non_existent_command_")
	result, ok := cliCommand.Run()

	assert.False(t, ok)
	assert.Equal(t, []byte("Command failed to run"), result)
}

func TestGetCommand(t *testing.T) {
	cliCommand := newCommandWithBinary([]string{"/"}, "ls")
	cmd := cliCommand.getCommand()

	assert.Equal(t, []string{"ls", "/"}, cmd.Args)
}
