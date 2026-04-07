// Package cmd implements the CLI commands using cobra.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	appIDFlag      string
	secretCodeFlag string
	baseURLFlag    string
	verboseFlag    bool
)

var rootCmd = &cobra.Command{
	Use:     "xparse-cli",
	Short:   "Textin xParser CLI — parse documents for Agents",
	Version: version,
	Long: `Textin xParser CLI is a command-line tool for document parsing powered by Textin xParser API.
Designed as Agent infrastructure — zero config, stdout-friendly, structured errors.

Supports: PDF, Images (png, jpg, bmp, tiff, webp), Doc(x), Ppt(x), Xls(x), HTML, TXT, OFD, RTF

  # Zero config — free API, markdown to stdout
  xparse-cli parse report.pdf

  # JSON view
  xparse-cli parse report.pdf --view json

  # Save to directory
  xparse-cli parse report.pdf --output ./result/

  # Specific pages
  xparse-cli parse report.pdf --page-range "1-5"

  # Batch from file list
  xparse-cli parse --list files.txt --output ./result/

  # Use paid API
  xparse-cli parse report.pdf --api paid

For more information, visit https://www.textin.com`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&appIDFlag, "app-id", "", "Textin App ID (overrides env and config)")
	rootCmd.PersistentFlags().StringVar(&secretCodeFlag, "secret-code", "", "Textin Secret Code (overrides env and config)")
	rootCmd.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "API base URL (for private deployments)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose mode, print HTTP details")
}
