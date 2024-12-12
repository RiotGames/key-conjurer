package command

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
)

type ShellType = string

const (
	shellTypePowershell ShellType = "powershell"
	shellTypeBash       ShellType = "bash"
	shellTypeBasic      ShellType = "basic"
	shellTypeInfer      ShellType = "infer"
)

func getShellType() ShellType {
	pid := os.Getppid()
	parentProc, _ := ps.FindProcess(pid)
	name := strings.ToLower(parentProc.Executable())

	if strings.Contains(name, "bash") || strings.Contains(name, "zsh") || strings.Contains(name, "ash") {
		return shellTypeBash
	}

	if strings.Contains(name, "powershell") || strings.Contains(name, "pwsh") {
		return shellTypePowershell
	}

	if runtime.GOOS == "windows" {
		return shellTypeBasic
	}

	return shellTypeBash
}

type CloudCredentials struct {
	AccountID       string `json:"AccountId"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func LoadAWSCredentialsFromEnvironment() CloudCredentials {
	return CloudCredentials{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		AccountID:       os.Getenv("AWSKEY_ACCOUNT"),
		Expiration:      os.Getenv("AWSKEY_EXPIRATION"),
	}
}

func (c *CloudCredentials) ValidUntil(account *Account, dur time.Duration) bool {
	if account == nil || c == nil {
		return false
	}

	if c.AccountID != account.ID {
		return false
	}

	expiration, err := time.Parse(time.RFC3339, c.Expiration)
	if err != nil {
		return false
	}

	return expiration.After(time.Now().Add(dur))
}

type bashWriter struct{}

func (bashWriter) ExportEnvironmentVariable(w io.Writer, key, value string) (int, error) {
	return fmt.Fprintf(w, "export %s=%s\n", key, value)
}

type powershellWriter struct{}

func (powershellWriter) ExportEnvironmentVariable(w io.Writer, key, value string) (int, error) {
	return fmt.Fprintf(w, "$Env:%s = %q\r\n", key, value)
}

type basicWriter struct{}

func (basicWriter) ExportEnvironmentVariable(w io.Writer, key, value string) (int, error) {
	return fmt.Fprintf(w, "SET %s=%s\r\n", key, value)
}

type environmentVariableWriter interface {
	ExportEnvironmentVariable(w io.Writer, key, value string) (int, error)
}

func (c CloudCredentials) WriteFormat(w io.Writer, format ShellType) (int, error) {
	var writer environmentVariableWriter
	if format == shellTypeInfer {
		format = getShellType()
	}

	switch format {
	case shellTypePowershell:
		writer = powershellWriter{}
	case shellTypeBasic:
		writer = basicWriter{}
	case shellTypeBash:
		writer = bashWriter{}
	}

	writer.ExportEnvironmentVariable(w, "TF_VAR_access_key", c.AccessKeyID)
	writer.ExportEnvironmentVariable(w, "TF_VAR_secret_key", c.SecretAccessKey)
	writer.ExportEnvironmentVariable(w, "TF_VAR_token", c.SessionToken)
	writer.ExportEnvironmentVariable(w, "AWSKEY_EXPIRATION", c.Expiration)
	writer.ExportEnvironmentVariable(w, "AWSKEY_ACCOUNT", c.AccountID)
	writer.ExportEnvironmentVariable(w, "AWS_ACCESS_KEY_ID", c.AccessKeyID)
	writer.ExportEnvironmentVariable(w, "AWS_SECRET_ACCESS_KEY", c.SecretAccessKey)
	writer.ExportEnvironmentVariable(w, "AWS_SESSION_TOKEN", c.SessionToken)
	writer.ExportEnvironmentVariable(w, "AWS_SECURITY_TOKEN", c.SessionToken)

	return 0, nil
}
