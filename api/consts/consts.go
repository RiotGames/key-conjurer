package consts

import (
	"time"
)

var (
	HttpTimeout      time.Duration = 120 * time.Second
	Version                        = "go run"
	LogstashEndpoint               = ""
)
