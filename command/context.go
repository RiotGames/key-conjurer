package command

import (
	"context"

	"github.com/spf13/cobra"
)

type configInfo struct {
	Path   string
	Config *Config
}

type ctxKeyConfig struct{}

func ConfigFromCommand(cmd *cobra.Command) *Config {
	return cmd.Context().Value(ctxKeyConfig{}).(*configInfo).Config
}

func ConfigPathFromCommand(cmd *cobra.Command) string {
	return cmd.Context().Value(ctxKeyConfig{}).(*configInfo).Path
}

func ConfigContext(ctx context.Context, config *Config, path string) context.Context {
	return context.WithValue(ctx, ctxKeyConfig{}, &configInfo{Path: path, Config: config})
}
