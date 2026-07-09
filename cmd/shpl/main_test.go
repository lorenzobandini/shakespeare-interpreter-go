package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMainEngineSetup(t *testing.T) {
	expectedStatus := true

	if !expectedStatus {
		t.Errorf("Setup validation failed: expected %t, got false", expectedStatus)
	}
}

func TestTokensCommand(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tokens", "../../testdata/lexer/minimal.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "{WORD") {
		t.Errorf("output missing WORD tokens:\n%s", output)
	}
	if !strings.Contains(output, "{EOF") {
		t.Errorf("output missing EOF token:\n%s", output)
	}
}
