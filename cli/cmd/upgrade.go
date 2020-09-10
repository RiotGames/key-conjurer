package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Downloads the latest version of keyconjurer.",
	Long:    "Downloads the latest version of keyconjurer.",
	Args:    cobra.ExactArgs(0),
	Example: "keyconjurer upgrade",
	Run: func(cmd *cobra.Command, args []string) {
		keyConjurerRcPath, err := os.Executable()

		if err != nil {
			log.Fatal(err)
		}

		switch runtime.GOOS {
		case ("windows"):
			windowsDownload(keyConjurerRcPath)
		default:
			defaultDownload(keyConjurerRcPath)
		}
	}}

// windowsDownload uses a special way to replace the binary due to restrictions in Windows. Because
//  you cannot replace the currently executing binary, a temporary script is created. This script
//  waits 3 seconds for the current process to exit, then downloads the latest Windows binary and
//  replaces the old one, finally it removes itself from the filesystem. The cmd prompt should
//  appear on the users screen to give them feedback that the download process began an ended.
func windowsDownload(keyConjurerRcPath string) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "keyconjurer-downloader-*.cmd")
	if err != nil {
		log.Fatalf("Unable to create download script. Reason: %v", err)
	}
	command := fmt.Sprintf("timeout 3 && bitsadmin /transfer keyconjurerdownload /priority foreground /download %s/%s %s && del %s && exit", keyconjurer.DownloadURL, keyconjurer.WindowsBinaryName, keyConjurerRcPath, tmpFile.Name())
	fileData := []byte(command)

	if _, err = tmpFile.Write(fileData); err != nil {
		log.Fatal("Failed to write to temporary file", err.Error())
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err.Error())
	}
	cmd := exec.Command("cmd", "/C", "start", tmpFile.Name())
	cmd.Start()
}

// defaultDownload replaces the currently executing binary by writing over it directly.
//  Gotta love how easy this is in Linux and OSX <3
func defaultDownload(keyConjurerRcPath string) {
	binary, err := keyconjurer.GetLatestBinary()
	if err != nil {
		log.Fatal("Unable to download the latest binary.")
	}

	if err := ioutil.WriteFile(keyConjurerRcPath, binary, 0744); err != nil {
		log.Fatalf("Could Not Save Binary. Reason: %v", err)
	}
}
