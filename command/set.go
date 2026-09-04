package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var cfgCmd = &cli.Command{
	Name:  "config",
	Usage: "Commands to manage KeyConjurer's configuration",
	Commands: []*cli.Command{
		setCmd,
		pathCmd,
	},
}

var setCmd = &cli.Command{
	Name:  "set",
	Usage: "Set configuration values",
	Commands: []*cli.Command{
		setTTLCmd,
		setTimeRemainingCmd,
	},
}

var pathCmd = &cli.Command{
	Name:  "path",
	Usage: "Print the absolute path to the configuration file",
	Action: func(_ context.Context, cmd *cli.Command) error {
		path, err := findConfigPath()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.Writer, path)
		return nil
	},
}

var setTTLCmd = &cli.Command{
	Name:  "ttl",
	Usage: "Sets ttl value in number of hours. Keys are requested to last at least this long.",
	Arguments: []cli.Argument{
		&cli.UintArg{Name: "ttl"},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		hours := cmd.UintArg("ttl")
		if hours == 0 {
			return cli.Exit("No ttl provided", 1)
		}

		if hours > 8 || hours < 1 {
			return cli.Exit("ttl must be between 1 and 8 hours", 1)
		}

		config := ConfigFromContext(ctx)
		config.TTL = hours
		return nil
	},
}

var setTimeRemainingCmd = &cli.Command{
	Name:  "time-remaining",
	Usage: "Sets time remaining value in number of minutes. Keys will be refreshed if they expire within this time period.",
	Arguments: []cli.Argument{
		&cli.UintArg{Name: "time-remaining"},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		timeRemaining := cmd.UintArg("time-remaining")
		if timeRemaining == 0 {
			return cli.Exit("No time remaining provided", 1)
		}
		config := ConfigFromContext(ctx)
		config.TimeRemaining = timeRemaining
		return nil
	},
}
