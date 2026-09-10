package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeFlagSet is a minimal flagSet backed by plain maps, used to test
// GetCommand.Parse in isolation without going through cli.Command.Run (which
// would require a full context.Config and risks touching the network/
// keychain once Action reaches Execute).
type fakeFlagSet struct {
	strings map[string]string
	bools   map[string]bool
	uints   map[string]uint
}

func (f fakeFlagSet) String(name string) string { return f.strings[name] }
func (f fakeFlagSet) Bool(name string) bool      { return f.bools[name] }
func (f fakeFlagSet) Uint(name string) uint      { return f.uints[name] }

// This guards the flag-name mismatch fixed alongside it: GetCommand.Parse
// must read back the exact flag names registered on getCmd ("output-type",
// "shell-type", "awscli-path", "role-name"), not the shared
// FlagOutputType/FlagShellType/FlagAWSCLIPath constants declared in
// switch.go (which hold "output"/"shell"/"awscli" - the names switch.go's
// own, differently-named flags use). Reading the wrong name silently
// produces an empty string, which fails Validate with a confusing
// "invalid output type" error regardless of what the user actually passed.
func TestGetCommandParseReadsItsOwnFlagNames(t *testing.T) {
	flags := fakeFlagSet{
		strings: map[string]string{
			"output-type": "json",
			"shell-type":  "bash",
			"role-name":   "SomeRole",
			"awscli-path": "/custom/aws/path",
		},
	}

	var g GetCommand
	err := g.Parse(flags, []string{"my-account"})
	assert.NoError(t, err)

	assert.Equal(t, "json", g.OutputType)
	assert.Equal(t, "bash", g.ShellType)
	assert.Equal(t, "SomeRole", g.RoleName)
	assert.Equal(t, "/custom/aws/path", g.AWSCLIPath)
	assert.Equal(t, "my-account", g.AccountIDOrName)

	assert.NoError(t, g.Validate())
}

func TestGetCommandParseRequiresAccount(t *testing.T) {
	var g GetCommand
	err := g.Parse(fakeFlagSet{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account name or alias is required")
}
