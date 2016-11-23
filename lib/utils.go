package lib

import (
	"os/exec"
	"fmt"
	"strings"
	"io/ioutil"
)

func PrintCmd(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func RunCmd(cmd *exec.Cmd) ([]byte, error) {
	PrintCmd(cmd)

	output, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	defer output.Close()
	defer stderr.Close()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	stdoutData, err := ioutil.ReadAll(output)
	if err != nil {
		return nil, err
	}

	stderrData, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	stdoutData = append(stdoutData[:], stderrData[:]...)

	cmd.Wait()
	return stdoutData, err
}