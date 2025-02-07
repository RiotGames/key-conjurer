package command

import (
	"context"

	"github.com/spf13/cobra"
)

type ctxKeyConfig struct{}

func ConfigFromCommand(cmd *cobra.Command) *Config {
	return cmd.Context().Value(ctxKeyConfig{}).(*Config)
}

func ConfigContext(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, ctxKeyConfig{}, config)
}
