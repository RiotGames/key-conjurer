package main

// Vars for build time
var (
	ClientID       string
	OIDCDomain     string
	ServerAddress  string
	Version        = "TBD"
	BuildTimestamp = "BuildTimestamp is not set"
	// CallbackPorts is a list of ports that will be attempted in no particular order for hosting an Oauth2 callback web server.
	// This cannot be set using -ldflags='-X ..' because -X requires that this be a string literal or uninitialized.
	//
	// These ports are chosen somewhat arbitrarily
	CallbackPorts = []string{"57468", "47512", "57123", "61232", "48231", "49757", "59834", "54293"}
)

const (
	// DefaultTTL for requested credentials in hours
	DefaultTTL uint = 1
	// DefaultTimeRemaining for new key requests in minutes
	DefaultTimeRemaining uint = 5
)
