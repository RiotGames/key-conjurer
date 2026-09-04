package command

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func execute(cmd *cli.Command, args ...string) (string, error) {
	c := *cmd
	var buf bytes.Buffer
	c.Writer = &buf
	_ = c.Run(context.TODO(), append([]string{"keyconjurer"}, args...))
	return buf.String(), nil
}

func TestVersionFlag(t *testing.T) {
	output, err := execute(rootCmd, "--version")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := fmt.Sprintf("keyconjurer-%s-%s TBD (BuildTimestamp is not set)\n", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, output, expected)
}

func TestVersionShortFlag(t *testing.T) {
	output, err := execute(rootCmd, "-v")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := fmt.Sprintf("keyconjurer-%s-%s TBD (BuildTimestamp is not set)\n", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, output, expected)
}
