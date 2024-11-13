package command

import "context"

type AliasCommand struct {
	AccountName string `arg:""`
	Alias       string `arg:""`
}

func (a AliasCommand) Run(globals *Globals, config *Config) error {
	return a.RunContext(context.Background(), globals, config)
}

func (a AliasCommand) RunContext(ctx context.Context, _ *Globals, config *Config) error {
	config.Alias(a.AccountName, a.Alias)
	return nil
}

type UnaliasCommand struct {
	Alias string `arg:""`
}

func (a UnaliasCommand) Run(globals *Globals, config *Config) error {
	return a.RunContext(context.Background(), globals, config)
}

func (a UnaliasCommand) RunContext(ctx context.Context, _ *Globals, config *Config) error {
	config.Unalias(a.Alias)
	return nil
}
