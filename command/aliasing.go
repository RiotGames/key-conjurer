package command

import (
	"context"

	"github.com/urfave/cli/v3"
)

var aliasCmd = &cli.Command{
	Name:  "alias",
	Usage: "Alias an account to a nickname so you can refer to the account by the nickname.",
	Arguments: []cli.Argument{
		&cli.StringArg{Name: "accountName", Config: cli.StringConfig{TrimSpace: true}},
		&cli.StringArg{Name: "alias", Config: cli.StringConfig{TrimSpace: true}},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		config := ConfigFromContext(ctx)
		name := cmd.StringArg("accountName")
		alias := cmd.StringArg("alias")
		if name == "" {
			return cli.Exit("accountName is required", 1)
		}
		if alias == "" {
			return cli.Exit("alias is required", 1)
		}

		config.Alias(name, alias)
		return nil
	},
}

var unaliasCmd = &cli.Command{
	Name:  "unalias",
	Usage: "Remove alias from account.",
	Arguments: []cli.Argument{
		&cli.StringArg{Name: "alias", Config: cli.StringConfig{TrimSpace: true}},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		config := ConfigFromContext(ctx)
		alias := cmd.StringArg("alias")
		if alias == "" {
			return cli.Exit("alias is required", 1)
		}
		config.Unalias(alias)
		return nil
	},
}
