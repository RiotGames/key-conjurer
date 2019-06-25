package keyconjurer

// Vars for build time
var Version string = "go run"
var Client string = "go runtime"

// Var for switching APIs
var Dev bool = false

// DefaultTTL for requested credentials in hours
const DefaultTTL uint = 1

// DefaultTimeRemaining for new key requests in minutes
const DefaultTimeRemaining uint = 60

// available API  endpoints
var DevAPI string
var ProdAPI string
var DownloadURL string
