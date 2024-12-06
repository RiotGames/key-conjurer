package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kong"
)

type Globals struct {
	OIDCDomain string `help:"The domain name of your OIDC server." hidden:"" env:"KEYCONJURER_OIDC_DOMAIN" default:"${oidc_domain}"`
	ClientID   string `help:"The client ID of your OIDC server." hidden:"" env:"KEYCONJURER_CLIENT_ID" default:"${client_id}"`
	Quiet      bool   `help:"Tells the CLI to be quiet; stdout will not contain human-readable informational messages."`
}

type CLI struct {
	Globals

	Login    LoginCommand    `cmd:"" help:"Authenticate with KeyConjurer."`
	Accounts AccountsCommand `cmd:"" help:"Display accounts."`
	Get      GetCommand      `cmd:"" help:"Retrieve temporary cloud credentials."`
	Switch   SwitchCommand   `cmd:"" help:"Switch between accounts."`
	Roles    RolesCommand    `cmd:"" help:"Display roles for a specific account."`
	Alias    AliasCommand    `cmd:"" help:"Create an alias for an account."`
	Unalias  UnaliasCommand  `cmd:"" help:"Remove an alias."`
	Set      SetCommand      `cmd:"" help:"Set config values."`

	Version kong.VersionFlag `help:"Show version information." short:"v" flag:""`

	Config Config `kong:"-"`
}

func (CLI) Help() string {
	return `KeyConjurer retrieves temporary credentials from Okta with the assistance of an optional API.

To get started run the following commands:
  keyconjurer login
  keyconjurer accounts
  keyconjurer get <accountName>`
}

func getConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		// UserConfigDir() and UserHomeDir() do slightly different things. If UserConfigDir() fails, try UserHomeDir()
		cfgDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(cfgDir, "keyconjurer", "config.json"), nil
}

func (c *CLI) BeforeApply(ctx *kong.Context, trace *kong.Path) error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	file, err := EnsureConfigFileExists(cfgPath)
	if err != nil {
		return err
	}

	err = c.Config.Read(file)
	if err != nil {
		return err
	}

	// Make *Config available to all sub-commands.
	// This must be &c.Config because c.Config is not a pointer.
	ctx.Bind(&c.Config)
	return nil
}

func (c *CLI) AfterRun(ctx *kong.Context) error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Do not use EnsureConfigFileExists here! EnsureConfigFileExists opens the file in append mode.
	// If we open the file in append mode, we'll always append to the file. If we open the file in truncate mode before reading from the file, the content will be truncated _before we read from it_, which will cause a users configuration to be discarded every time we run the program.
	if err := os.MkdirAll(filepath.Dir(cfgPath), os.ModeDir|os.FileMode(0755)); err != nil {
		return err
	}

	file, err := os.Create(cfgPath)
	if err != nil {
		return fmt.Errorf("unable to create %s reason: %w", cfgPath, err)
	}

	defer file.Close()
	return c.Config.Write(file)
}

func newKong(cli *CLI) (*kong.Kong, error) {
	return kong.New(
		cli,
		kong.Name("keyconjurer"),
		kong.Description("Retrieve temporary cloud credentials."),
		kong.UsageOnError(),
		kong.Vars{
			"client_id":      ClientID,
			"server_address": ServerAddress,
			"oidc_domain":    OIDCDomain,
			"version":        fmt.Sprintf("keyconjurer-%s-%s %s (%s)", runtime.GOOS, runtime.GOARCH, Version, BuildTimestamp),
		},
	)
}

func Execute(args []string) error {
	var cli CLI
	k, err := newKong(&cli)
	if err != nil {
		return err
	}

	kongCtx, err := k.Parse(args)
	if err != nil {
		return err
	}

	return kongCtx.Run(&cli.Globals)
}
