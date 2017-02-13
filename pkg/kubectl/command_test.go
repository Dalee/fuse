package kubectl

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
	"os"
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
	os.Setenv("CLUSTER_CONTEXT", "live-context")
	cliCommand := newCommand([]string{
		"hello",
		"world",
	})

	cmdString := strings.Join(cliCommand.getCommand().Args, " ")
	assert.Equal(t, "kubectl --context=live-context hello world", cmdString)
	os.Unsetenv("CLUSTER_CONTEXT")
}

// ensure command can be executed
func TestExecuteCommand(t *testing.T) {
	cliCommand := newCommandWithBinary([]string{"/"}, "ls")
	_, ok := cliCommand.Run()
	assert.True(t, ok)
}
