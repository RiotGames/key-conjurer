package main

// Vars for build time
var Version string = "TBD"
var ClientName string = appname

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
const LinuxAmd64BinaryName string = "keyconjurer-linux-amd64"
const LinuxArm64BinaryName string = "keyconjurer-linux-arm64"
const WindowsBinaryName string = "keyconjurer-windows.exe"
const DarwinArm64BinaryName string = "keyconjurer-darwin-arm64"
const DarwinAmd64BinaryName string = "keyconjurer-darwin-amd64"

const appname string = "keyconjurer"
