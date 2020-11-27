package keyconjurer

import (
	"time"
)

// Vars for build time
var Version string = "go run"
var ClientName string = "go runtime"

// Var for switching APIs
var Dev bool = false

// DefaultTTL for requested credentials in hours
const DefaultTTL uint = 1

// DefaultTimeRemaining for new key requests in minutes
const DefaultTimeRemaining uint = 60

// available API  endpoints
var DownloadURL string

// CLI binary names
const LinuxBinaryName string = "keyconjurer-linux"
const WindowsBinaryName string = "keyconjurer-windows.exe"
const DarwinBinaryName string = "keyconjurer-darwin"

// CLI HTTP Timeouts
var ClientHttpTimeoutInSeconds time.Duration = 30
