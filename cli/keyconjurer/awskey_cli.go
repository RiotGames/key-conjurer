package keyconjurer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	homedir "github.com/mitchellh/go-homedir"
	ps "github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Login prompts the user for the AD credentials and then passes back the list
//  of AWS applications and encrypted creds via the inputed userData
func Login(keyConjurerRcPath string, prompt bool) *UserData {
	shouldPrompt := prompt

	expandedKeyConjurerRcPath, err := homedir.Expand(keyConjurerRcPath)
	if err != nil {
		Logger.Errorln(err)
		Logger.Fatalf("Can't expand location %v", keyConjurerRcPath)

		// do we want to exit if we cant expand the path???
	}

	touchFileIfNotExist(expandedKeyConjurerRcPath)
	userData := &UserData{
		filePath: expandedKeyConjurerRcPath,
	}

	if err := userData.LoadFromFile(); err != nil {
		Logger.Warnf("unable to load keyconjurerrc at location %s\n", expandedKeyConjurerRcPath)
		Logger.Warnln(err)
		userData.SetDefaults()
		shouldPrompt = true
	}

	if shouldPrompt {
		err = userData.promptForADCreds()

		if err != nil {
			Logger.Error("error using provided AD credentials")
			Logger.Fatal(err)

		}

		err = userData.Save()

		if err != nil {
			Logger.Error("error saving user data")
			Logger.Fatal(err)
		}
		fmt.Println("Successfully logged in")
	}
	return userData

}

// GetUsernameAndPassword prompts the user for their username and password via stdin
func getUsernameAndPassword() (string, string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("username: ")
	username := ""
	if scanner.Scan() {
		username = scanner.Text()
	} else {
		return "", "", errors.New("Unable to get username")
	}

	fmt.Printf("password: ")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		Logger.Warn("Unable to get password")
		return "", "", errors.New("Unable to get password")
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

func touchFileIfNotExist(path string) {
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		errmkdir := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.FileMode(uint32(0755)))
		if errmkdir != nil {
			Logger.Fatal(errmkdir)
		}
	} else if err != nil {
		Logger.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, errCreate := os.Create(path)
		if errCreate != nil {
			Logger.Fatal(err)
		}
		f.Close()
	} else if err != nil {
		Logger.Fatal(err)
	}
}
