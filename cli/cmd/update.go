package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/textin/xparser-ecosystem/cli/internal/output"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update xParser CLI to the latest version",
	Long:  `Check for updates and download the latest version of the xParser CLI binary.`,
	RunE:  runUpdate,
}

var (
	updateCheckOnly bool
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Only check for updates, don't download")
}

// ReleaseInfo represents a GitHub/CDN release entry.
type ReleaseInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	output.Status("Current version: %s", version)

	if version == "dev" {
		return fmt.Errorf("cannot update a dev build. Please build from source or download a release")
	}

	output.Status("Checking for updates...")

	latest, err := checkLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if latest.Version == version {
		output.Status("Already up to date")
		return nil
	}

	output.Status("New version available: %s -> %s", version, latest.Version)

	if updateCheckOnly {
		return nil
	}

	downloadURL := latest.URL
	if downloadURL == "" {
		return fmt.Errorf("no download URL available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	output.Status("Downloading %s...", downloadURL)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	tmpPath := execPath + ".new"
	if err := downloadFile(downloadURL, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod failed: %w", err)
	}

	oldPath := execPath + ".old"
	os.Remove(oldPath)

	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		// Try to restore
		os.Rename(oldPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	os.Remove(oldPath)
	output.Status("Updated successfully to %s", latest.Version)
	return nil
}

func checkLatestVersion() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// This URL should be replaced with the actual release endpoint
	url := "https://api.textin.com/xparser-cli/releases/latest"
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &info, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
