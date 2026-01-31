package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	// RuleEnginePath is the installed location of the Haskell rule engine
	RuleEnginePath = "/usr/local/lib/kubecheck/kubecheck-rules"
)

// callRuleEngine invokes the Haskell rule engine with JSON input
func callRuleEngine(input []byte) ([]byte, error) {
	cmd := exec.Command(RuleEnginePath)
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("rule engine execution failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
