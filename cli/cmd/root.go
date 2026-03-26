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
	Use:     "xparser",
	Short:   "Textin xParser CLI — turn documents into Markdown",
	Version: version,
	Long: `Textin xParser CLI is a command-line tool for document parsing powered by Textin xParser API.

Supports: PDF, Images (png, jpg, bmp, tiff, webp), Doc(x), Ppt(x), Xls(x), HTML, TXT, OFD, RTF

  # Parse a document to markdown
  xparser parse report.pdf                          # markdown to stdout
  xparser parse report.pdf -o ./out/                # save to directory
  xparser parse report.pdf --parse-mode vlm         # use VLM mode
  xparser parse *.pdf -o ./results/                  # batch mode

  # Parse with options
  xparser parse report.pdf --table-flavor md --get-image objects -o ./out/

Authenticate first:

  xparser auth

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
