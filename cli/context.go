package main

import "context"

type configInfo struct {
	Path   string
	Config *Config
}

type ctxKeyConfig struct{}

func ConfigFromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKeyConfig{}).(*configInfo).Config
}

func ConfigPathFromContext(ctx context.Context) string {
	return ctx.Value(ctxKeyConfig{}).(*configInfo).Path
}

func ConfigContext(ctx context.Context, config *Config, path string) context.Context {
	return context.WithValue(ctx, ctxKeyConfig{}, configInfo{Path: path, Config: config})
}
