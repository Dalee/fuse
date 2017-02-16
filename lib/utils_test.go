package lib

import (
	"os/exec"
	"testing"
)

func TestCommandRun(t *testing.T) {
	cmd := exec.Command("ls", "-lha", "/")
	output, success := RunCmd(cmd)

	if success != true {
		t.Error("Failed to execute")
	}

	if len(output) == 0 {
		t.Error("Failed to get output")
	}
}

func TestCommandRunError(t *testing.T) {
	cmd := exec.Command("ls", "-")
	output, success := RunCmd(cmd)

	if success != false {
		t.Error("Command executed, but it should'nt")
	}

	if len(output) == 0 {
		t.Error("Failed to get output")
	}
}
