package command

import (
	"context"
)

type ctxKeyConfig struct{}

func ConfigFromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKeyConfig{}).(*Config)
}

func ConfigContext(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, ctxKeyConfig{}, config)
}
