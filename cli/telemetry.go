package main

import (
	"io"
	"os"

	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func isTelemetryEnabled() bool {
	return os.Getenv("KEYCONJURER_TELEMETRY") != "off"
}

func loadNewRelicApplicationInfo() (newrelic.Application, bool) {
	if isTelemetryEnabled() {
		return newrelic.Application{}, false
	}

	return newrelic.Application{}, false
}

func NewTelemetryWriter(w io.Writer) io.Writer {
	if appInfo, ok := loadNewRelicApplicationInfo(); ok {
		return logWriter.New(w, &appInfo)
	}

	return w
}
