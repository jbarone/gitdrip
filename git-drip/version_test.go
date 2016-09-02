package main

import (
	"bytes"
	"testing"
)

func TestCmdVersion(t *testing.T) {
	stdoutTrap = new(bytes.Buffer)
	defer func() {
		stdoutTrap = nil
	}()

	cmdVersion()

	expected := version + "\n"

	if stdoutTrap.String() != expected {
		t.Errorf("Expected '%s' got: '%s'", expected, stdoutTrap.String())
	}
}

func TestVersion(t *testing.T) {
	testMain(t, "version")
	expected := version + "\n"

	if testStdout.String() != expected {
		t.Errorf("Expected '%s' got: '%s'", expected, testStdout.String())
	}
}
