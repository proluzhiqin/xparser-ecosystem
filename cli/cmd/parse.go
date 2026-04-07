package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
	"github.com/textin/xparser-ecosystem/cli/internal/exitcode"
)

// ── V1 parse flags ──

var (
	parseView              string // json | markdown
	parseAPI               string // free | paid
	parsePageRangeFlag     string // "1-5" or "1-2,5-10"
	parsePassword          string
	parseIncludeCharDetail bool
	parseListFile          string
	parseOutput            string
)

var parseCmd = &cobra.Command{
	Use:   "parse <file-or-url>",
	Short: "Parse a document to Markdown or JSON",
	Long: `Parse documents using the Textin xParser API.

Supports: PDF, Images (png, jpg, bmp, tiff, webp), Doc(x), Ppt(x), Xls(x),
          HTML, TXT, OFD, RTF

Examples:
  # Basic usage — markdown to stdout (zero config, uses free API)
  xparse-cli parse report.pdf

  # View as JSON
  xparse-cli parse report.pdf --view json

  # Save to file
  xparse-cli parse report.pdf --output ./result/

  # Parse specific pages
  xparse-cli parse report.pdf --page-range "1-5"

  # Encrypted PDF
  xparse-cli parse secret.pdf --password mypassword

  # Batch from file list
  xparse-cli parse --list files.txt --output ./result/

  # Use paid API explicitly
  xparse-cli parse report.pdf --api paid`,
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVar(&parseView, "view", "markdown", "Output view: markdown, json")
	parseCmd.Flags().StringVar(&parseAPI, "api", "", "API mode: free, paid (default: paid if key exists, else free)")
	parseCmd.Flags().StringVar(&parsePageRangeFlag, "page-range", "", `Page range, e.g. "1-5" or "1-2,5-10"`)
	parseCmd.Flags().StringVar(&parsePassword, "password", "", "Password for encrypted documents")
	parseCmd.Flags().BoolVar(&parseIncludeCharDetail, "include-char-details", false, "Include character-level details (coordinates, confidence)")
	parseCmd.Flags().StringVar(&parseListFile, "list", "", "Read input list from file (one path per line); requires --output")
	parseCmd.Flags().StringVarP(&parseOutput, "output", "o", "", "Output file path or directory; omit for stdout")
}

// ── entry point ──

func runParse(cmd *cobra.Command, args []string) error {
	// Validate --view
	if parseView != "markdown" && parseView != "json" {
		return usageErr(exitcode.ErrInvalidView)
	}

	// Validate --api
	var apiMode APIMode
	switch parseAPI {
	case "", "free", "paid":
		apiMode = APIMode(parseAPI)
	default:
		return usageErr(exitcode.ErrInvalidAPI)
	}

	// Collect input sources
	sources, err := collectSources(args, parseListFile)
	if err != nil {
		return usageErr(exitcode.ErrOpenListFile)
	}

	if len(sources) == 0 {
		return usageErr(exitcode.ErrNoInput)
	}

	// Validate file existence early
	for _, src := range sources {
		if !isURL(src) {
			if _, err := os.Stat(src); os.IsNotExist(err) {
				if looksLikeBoolValue(src) {
					return usageErr(exitcode.ErrFlagValueNotFile)
				}
				return usageErr(exitcode.ErrFileNotFound)
			}
		}
	}

	// --list requires --output
	if parseListFile != "" && parseOutput == "" {
		return usageErr(exitcode.ErrListRequiresOut)
	}

	// Batch mode requires --output to be a directory
	if len(sources) > 1 && parseOutput == "" {
		return usageErr(exitcode.ErrMultiRequiresOut)
	}

	// Resolve credentials
	credSrc, err := config.ResolveCredentials(cmd)
	if err != nil {
		return generalErr(exitcode.ErrCredentialsConfig)
	}

	// Determine API mode
	isFree := resolveAPIMode(apiMode, credSrc)

	// --api paid requires credentials
	if !isFree && (credSrc.AppID == "" || credSrc.SecretCode == "") {
		return usageErr(exitcode.ErrPaidNoCreds)
	}

	client := newXParserClient(cmd, credSrc, isFree)

	opts := &ParseOptions{
		PageRange:          parsePageRangeFlag,
		Password:           parsePassword,
		IncludeCharDetails: parseIncludeCharDetail,
	}

	if len(sources) == 1 {
		return runSingleParse(client, sources[0], opts)
	}
	return runBatchParse(client, sources, opts)
}

// ── single file/url ──

func runSingleParse(client *XParserClient, source string, opts *ParseOptions) error {
	var resp *ParseResponse
	var err error

	if isURL(source) {
		resp, err = client.ParseURL(source, opts)
	} else {
		resp, err = client.ParseFile(source, opts)
	}
	if err != nil {
		return generalErr(exitcode.ErrNetworkRequest)
	}

	if resp.Code != 200 {
		return apiErr(resp.Code)
	}

	if !resp.HasResult() {
		return generalErr(exitcode.ErrNoResultData)
	}

	return outputResult(resp, source)
}

// ── batch ──

func runBatchParse(client *XParserClient, sources []string, opts *ParseOptions) error {
	if err := os.MkdirAll(parseOutput, 0o755); err != nil {
		return generalErr(exitcode.ErrCreateOutputDir)
	}

	failed := 0

	for _, source := range sources {
		var resp *ParseResponse
		var err error

		if isURL(source) {
			resp, err = client.ParseURL(source, opts)
		} else {
			resp, err = client.ParseFile(source, opts)
		}

		if err != nil {
			failed++
			continue
		}

		if resp.Code != 200 {
			failed++
			continue
		}

		if !resp.HasResult() {
			failed++
			continue
		}

		if _, err := saveResult(resp, source, parseOutput); err != nil {
			failed++
			continue
		}
	}

	if failed > 0 {
		return generalErr(exitcode.ErrBatchPartial)
	}
	return nil
}

// ── output ──

func outputResult(resp *ParseResponse, source string) error {
	if parseOutput == "" {
		// stdout mode
		switch parseView {
		case "json":
			data, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(data))
		default: // "markdown"
			fmt.Print(resp.GetMarkdown())
		}
		return nil
	}

	// file mode
	if _, err := saveResult(resp, source, parseOutput); err != nil {
		return generalErr(exitcode.ErrSaveResult)
	}
	return nil
}

func saveResult(resp *ParseResponse, source string, outputPath string) (string, error) {
	dir, base := resolveOutputTarget(outputPath, source)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	ext := ".md"
	if parseView == "json" {
		ext = ".json"
	}

	outPath := filepath.Join(dir, base+ext)

	switch parseView {
	case "json":
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal json: %w", err)
		}
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return "", fmt.Errorf("failed to save json: %w", err)
		}
	default: // "markdown"
		if err := os.WriteFile(outPath, []byte(resp.GetMarkdown()), 0o644); err != nil {
			return "", fmt.Errorf("failed to save markdown: %w", err)
		}
	}

	return outPath, nil
}

// ── error helpers ──

// apiErr outputs "api_code：message" to stderr and returns exit code 3.
func apiErr(apiCode int) *exitError {
	info := exitcode.FromAPICode(apiCode)
	if info == nil {
		return nil
	}
	fmt.Fprintf(os.Stderr, "%d：%s\n", info.APICode, info.Message)
	return &exitError{code: exitcode.APIError, msg: info.Message}
}

// usageErr outputs plain text to stderr and returns exit code 2.
func usageErr(message string) *exitError {
	fmt.Fprintln(os.Stderr, message)
	return &exitError{code: exitcode.UsageError, msg: message}
}

// generalErr outputs plain text to stderr and returns exit code 1.
func generalErr(message string) *exitError {
	fmt.Fprintln(os.Stderr, message)
	return &exitError{code: exitcode.GeneralError, msg: message}
}

// exitError carries a process exit code alongside the error message.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

// ExitCode returns the process exit code for this error.
func (e *exitError) ExitCode() int { return e.code }

// ── shared helpers ──

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func collectSources(args []string, listFile string) ([]string, error) {
	var sources []string
	sources = append(sources, args...)

	if listFile != "" {
		f, err := os.Open(listFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open list file: %w", err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				sources = append(sources, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read list file: %w", err)
		}
	}

	return sources, nil
}

func baseNameNoExt(source string) string {
	base := filepath.Base(source)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

func resolveOutputTarget(outputPath, source string) (dir, base string) {
	dir = outputPath
	base = baseNameNoExt(source)

	// If outputPath is an existing file or looks like a file path, use it directly
	info, err := os.Stat(outputPath)
	if err == nil && !info.IsDir() {
		return filepath.Dir(outputPath), baseNameNoExt(outputPath)
	}

	if outputPathLooksLikeFile(outputPath) {
		return filepath.Dir(outputPath), baseNameNoExt(outputPath)
	}

	return dir, base
}

func outputPathLooksLikeFile(outputPath string) bool {
	if outputPath == "" || hasTrailingPathSeparator(outputPath) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	return ext == ".md" || ext == ".json" || ext == ".markdown"
}

func hasTrailingPathSeparator(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\")
}

func looksLikeBoolValue(s string) bool {
	switch strings.ToLower(s) {
	case "0", "1", "true", "false", "yes", "no":
		return true
	}
	return false
}
