package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Downloads the latest version of KeyConjurer",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyConjurerRcPath, err := os.Executable()
		if err != nil {
			return err
		}

		switch runtime.GOOS {
		case "windows":
			return windowsDownload(keyConjurerRcPath)
		default:
			return defaultDownload(context.Background(), NewHTTPClient(), keyConjurerRcPath)
		}
	}}

// windowsDownload uses a special way to replace the binary due to restrictions in Windows. Because
//
//	you cannot replace the currently executing binary, a temporary script is created. This script
//	waits 3 seconds for the current process to exit, then downloads the latest Windows binary and
//	replaces the old one, finally it removes itself from the filesystem. The cmd prompt should
//	appear on the users screen to give them feedback that the download process began an ended.
func windowsDownload(keyConjurerRcPath string) error {
	f, err := os.CreateTemp(os.TempDir(), "keyconjurer-downloader-*.cmd")
	if err != nil {
		return fmt.Errorf("unable to create download script: %w", err)
	}

	command := fmt.Sprintf("timeout 3 && bitsadmin /transfer keyconjurerdownload /priority foreground /download %s/%s %s && del %s && exit", DownloadURL, WindowsBinaryName, keyConjurerRcPath, f.Name())
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
func defaultDownload(ctx context.Context, client *http.Client, keyConjurerRcPath string) error {
	tmp, err := os.CreateTemp(os.TempDir(), "keyconjurer")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for upgrade: %w", err)
	}

	defer tmp.Close()
	src, err := DownloadLatestBinary(ctx, client, tmp)
	if err != nil {
		return fmt.Errorf("unable to download the latest binary: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("could not close tmp file: %w", err)
	}

	bytesCopied, err := io.Copy(tmp, src)
	if err != nil {
		return fmt.Errorf("failed to copy new keyconjurer: %s", err)
	}

	// Re-open the temporary file for reading and copy:
	r, err := os.Open(tmp.Name())
	if err != nil {
		return fmt.Errorf("could not open temporary file %s: %w", tmp.Name(), err)
	}

	kc, _ := os.OpenFile(keyConjurerRcPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		return fmt.Errorf("unable to open %q: %w", keyConjurerRcPath, err)
	}

	bytesCopied2, err := io.Copy(kc, r)
	if err != nil || bytesCopied != bytesCopied2 {
		// If an error occurs here, KeyConjurer has been overwritten and is potentially corrrupted
		return fmt.Errorf("failed to copy new keyconjurer contents - keyconjurer is potentially corrupted and may need to be downloaded again: %w", err)
	}

	return nil
}

func getBinaryName() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return LinuxArm64BinaryName
		}

		return LinuxAmd64BinaryName
	case "windows":
		return WindowsBinaryName
	default:
		if runtime.GOARCH == "arm64" {
			return DarwinArm64BinaryName
		}
		return DarwinAmd64BinaryName
	}
}

// DownloadLatestBinary downloads the latest keyconjurer binary from the web.
func DownloadLatestBinary(ctx context.Context, client *http.Client, w io.Writer) (io.ReadCloser, error) {
	binaryURL := fmt.Sprintf("%s/%s", DownloadURL, getBinaryName())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, binaryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not upgrade: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not upgrade: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, errors.New("could not upgrade: response did not indicate success - are you being blocked by the server?")
	}

	return req.Body, nil
}
