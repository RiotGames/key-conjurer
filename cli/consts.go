package main

import (
	"time"

	"github.com/riotgames/key-conjurer/api/keyconjurer"
)

// Vars for build time
var Version string = "TBD"
var ClientName string = "go runtime"

var BuildDate string = "date not set"
var BuildTime string = "time not set"
var BuildTimeZone string = "zone not set"

// Var for switching APIs
var Dev bool = false

// DefaultTTL for requested credentials in hours
const DefaultTTL uint = 1

// DefaultTimeRemaining for new key requests in minutes
const DefaultTimeRemaining uint = 5

// available API  endpoints
var DownloadURL string = "URL not set yet"

// CLI binary names
const LinuxBinaryName string = "keyconjurer-linux"
const WindowsBinaryName string = "keyconjurer-windows.exe"
const DarwinBinaryName string = "keyconjurer-darwin"

const appname string = "keyconjurer"

// CLI HTTP Timeouts
var ClientHttpTimeoutInSeconds time.Duration = 30

const defaultIdentityProvider = keyconjurer.AuthenticationProviderOkta
