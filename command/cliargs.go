package command

import "github.com/urfave/cli/v3"

// positionalArgs returns the value of the named positional cli.Argument as a
// single-element []string, or an empty slice if it was not provided.
//
// This exists because cli.Command.Arguments consumes matching positional
// values out of cmd.Args() before Action runs: once a command declares
// Arguments, reading cmd.Args().Slice() (or cmd.Args().First()) in its Action
// will always come back empty. The value must be read back via
// cmd.StringArg(name) instead. See github.com/urfave/cli/v3's
// command_run.go, where cmd.parsedArgs is reassigned to the leftover,
// already-consumed args once len(cmd.Arguments) > 0.
func positionalArgs(cmd *cli.Command, name string) []string {
	if v := cmd.StringArg(name); v != "" {
		return []string{v}
	}
	return nil
}
