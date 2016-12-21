package lib

import (
	"testing"
	"strings"
)

func TestFormatCommandWithCluster(t *testing.T) {

	cmd := CommandFactory("prod-context", []string{
		"apply",
		"-f",
		"kubernetes.yml",
	})

	cmdStr := strings.Join(cmd.Args, " ")
	if cmdStr != "kubectl --context=prod-context apply -f kubernetes.yml" {
		t.Error("Incorrect command: ", cmdStr)
	}
}

func TestFormatCommandWithoutCluster(t *testing.T) {

	cmd := CommandFactory("", []string{
		"apply",
		"-f",
		"kubernetes.yml",
	})

	cmdStr := strings.Join(cmd.Args, " ")
	if cmdStr != "kubectl apply -f kubernetes.yml" {
		t.Error("Incorrect command: ", cmdStr)
	}
}