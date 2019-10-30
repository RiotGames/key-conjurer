package consts

import (
	"time"
)

var Version string = "go run"

var AuthenticatorSelect string = "onelogin"

var MFASelect string = "duo"

var SettingsRetrieverSelect = "kms_blob"

var LogstashEndpoint string = ""

var HttpTimeoutInSeconds time.Duration = 30
