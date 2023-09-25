package main

// Vars for build time
var (
	ClientID       string
	OIDCDomain     string
	ServerAddress  string
	Version        = "TBD"
	BuildTimestamp = "BuildTimestamp is not set"
	DownloadURL    = "URL not set yet"
)

const (
	// DefaultTTL for requested credentials in hours
	DefaultTTL uint = 1
	// DefaultTimeRemaining for new key requests in minutes
	DefaultTimeRemaining  uint   = 5
	LinuxAmd64BinaryName  string = "keyconjurer-linux-amd64"
	LinuxArm64BinaryName  string = "keyconjurer-linux-arm64"
	WindowsBinaryName     string = "keyconjurer-windows.exe"
	DarwinArm64BinaryName string = "keyconjurer-darwin-arm64"
	DarwinAmd64BinaryName string = "keyconjurer-darwin-amd64"
)
