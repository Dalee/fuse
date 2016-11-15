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

	defer output.Close()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(output)
	if err != nil {
		return nil, err
	}

	cmd.Wait()
	return data, err
}