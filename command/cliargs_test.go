package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

// This documents and guards the exact bug fixed alongside it: a cli.Command
// that declares Arguments must read them back via cmd.StringArg(name), not
// cmd.Args().Slice()/.First(), because cli.Command.Arguments consumes
// matching positional values out of Args() before Action runs. get.go,
// switch.go and roles.go all made this mistake when they were written
// against urfave/cli v3; alias/unalias never did, which is how the bug went
// unnoticed. See positionalArgs' doc comment in cliargs.go.
func TestPositionalArgsSurvivesArgumentsConsumption(t *testing.T) {
	var gotViaArgs, gotViaHelper string

	cmd := &cli.Command{
		Name: "probe",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "account"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			gotViaArgs = cmd.Args().First()
			gotViaHelper = firstOrEmpty(positionalArgs(cmd, "account"))
			return nil
		},
	}

	err := cmd.Run(context.Background(), []string{"probe", "my-account"})
	assert.NoError(t, err)

	assert.Empty(t, gotViaArgs, "cmd.Args().First() should be empty: the Arguments field already consumed it")
	assert.Equal(t, "my-account", gotViaHelper, "positionalArgs(cmd, name) should recover the value cmd.Args() lost")
}

func firstOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}
