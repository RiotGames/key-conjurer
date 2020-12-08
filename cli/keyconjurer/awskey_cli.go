package keyconjurer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	homedir "github.com/mitchellh/go-homedir"
	ps "github.com/mitchellh/go-ps"
)

// Login prompts the user for the AD credentials and then passes back the list
//  of AWS applications and encrypted creds via the inputed userData
func Login(keyConjurerRcPath string, prompt bool) (*UserData, error) {
	reader, err := openKeyConjurerRc(keyConjurerRcPath)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	var userData UserData
	if err := userData.Load(reader); err != nil {
		userData.SetDefaults()
		prompt = true
	}

	if !prompt {
		return &userData, nil
	}

	if err := userData.promptForADCreds(); err != nil {
		return nil, fmt.Errorf("error using provided AD credentials: %w", err)
	}

	if err := userData.Save(); err != nil {
		return nil, fmt.Errorf("error saving user data: %w", err)
	}

	// Don't like this, nope
	fmt.Println("Successfully logged in")
	return &userData, nil
}

var errUnableToReadUsername = errors.New("unable to read username")

// getUsernameAndPassword prompts the user for their username and password via stdin
func getUsernameAndPassword() (string, string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("username: ")
	username := ""
	if scanner.Scan() {
		username = scanner.Text()
	} else {
		return "", "", errUnableToReadUsername
	}

	fmt.Printf("password: ")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("unable to get password: %w", err)
	}

	password := string(bytes)
	// Need to add our own newline
	fmt.Println()
	return username, password, nil
}

func getShellType() string {
	pid := os.Getppid()
	parentProc, _ := ps.FindProcess(pid)
	normalizedName := strings.ToLower(parentProc.Executable())

	if strings.Contains(normalizedName, "powershell") || strings.Contains(normalizedName, "pwsh") {
		return "powershell"
	}
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return normalizedName
}

func openKeyConjurerRc(path string) (io.ReadCloser, error) {
	// It is probably an error if keyconjurerrc can't be read from
	expanded, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("can't expand location %s", path)
	}

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
		return nil, err
	}

	return os.OpenFile(expanded, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}
