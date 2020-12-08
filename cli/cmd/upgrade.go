package cmd

import (
	"fmt"
	"io/ioutil"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		keyConjurerRcPath, err := os.Executable()
		if err != nil {
			return err
		}

		switch runtime.GOOS {
		case "windows":
			return windowsDownload(keyConjurerRcPath)
		default:
			return defaultDownload(keyConjurerRcPath)
		}
	}}

// windowsDownload uses a special way to replace the binary due to restrictions in Windows. Because
//  you cannot replace the currently executing binary, a temporary script is created. This script
//  waits 3 seconds for the current process to exit, then downloads the latest Windows binary and
//  replaces the old one, finally it removes itself from the filesystem. The cmd prompt should
//  appear on the users screen to give them feedback that the download process began an ended.
func windowsDownload(keyConjurerRcPath string) error {
	f, err := ioutil.TempFile(os.TempDir(), "keyconjurer-downloader-*.cmd")
	if err != nil {
		return fmt.Errorf("unable to create download script: %w", err)
	}

	command := fmt.Sprintf("timeout 3 && bitsadmin /transfer keyconjurerdownload /priority foreground /download %s/%s %s && del %s && exit", keyconjurer.DownloadURL, keyconjurer.WindowsBinaryName, keyConjurerRcPath, f.Name())
	fileData := []byte(command)

	if _, err = f.Write(fileData); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := f.Close(); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/C", "start", f.Name())
	return cmd.Start()
}

// defaultDownload replaces the currently executing binary by writing over it directly.
//  Gotta love how easy this is in Linux and OSX <3
func defaultDownload(keyConjurerRcPath string) error {
	binary, err := keyconjurer.GetLatestBinary()
	if err != nil {
		return fmt.Errorf("unable to download the latest binary: %w", err)
	}

	if err := ioutil.WriteFile(keyConjurerRcPath, binary, 0744); err != nil {
		return fmt.Errorf("could not save binary: %w", err)
	}

	return nil
}
