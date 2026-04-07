package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
	"github.com/textin/xparser-ecosystem/cli/internal/exitcode"
	"github.com/textin/xparser-ecosystem/cli/internal/output"
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
		return &exitError{code: exitcode.UsageError, msg: fmt.Sprintf("invalid --view %q: must be 'markdown' or 'json'", parseView)}
	}

	// Validate --api
	var apiMode APIMode
	switch parseAPI {
	case "", "free", "paid":
		apiMode = APIMode(parseAPI)
	default:
		return &exitError{code: exitcode.UsageError, msg: fmt.Sprintf("invalid --api %q: must be 'free' or 'paid'", parseAPI)}
	}

	// Collect input sources
	sources, err := collectSources(args, parseListFile)
	if err != nil {
		return &exitError{code: exitcode.UsageError, msg: err.Error()}
	}

	if len(sources) == 0 {
		return &exitError{code: exitcode.UsageError, msg: "no input files specified. Provide files as arguments or use --list <file>"}
	}

	// Validate file existence early
	for _, src := range sources {
		if !isURL(src) {
			if _, err := os.Stat(src); os.IsNotExist(err) {
				// Detect likely misuse of bool flags (e.g. --include-char-details 1)
				if looksLikeBoolValue(src) {
					return &exitError{
						code: exitcode.UsageError,
						msg:  fmt.Sprintf("file not found: %q — this looks like a flag value, not a file. Bool flags don't take separate values; use --include-char-details (no value) or --include-char-details=true", src),
					}
				}
				return &exitError{code: exitcode.UsageError, msg: fmt.Sprintf("file not found: %s", src)}
			}
		}
	}

	// --list requires --output
	if parseListFile != "" && parseOutput == "" {
		return &exitError{code: exitcode.UsageError, msg: "--list requires --output to specify output directory"}
	}

	// Batch mode requires --output to be a directory
	if len(sources) > 1 && parseOutput == "" {
		return &exitError{code: exitcode.UsageError, msg: "multiple inputs require --output to specify output directory"}
	}

	// Resolve credentials
	credSrc, err := config.ResolveCredentials(cmd)
	if err != nil {
		return &exitError{code: exitcode.GeneralError, msg: err.Error()}
	}

	// Determine API mode
	isFree := resolveAPIMode(apiMode, credSrc)

	// --api paid requires credentials
	if !isFree && (credSrc.AppID == "" || credSrc.SecretCode == "") {
		return &exitError{
			code: exitcode.UsageError,
			msg:  "paid API requires credentials. Set XPARSE_APP_ID and XPARSE_SECRET_CODE, or run 'xparse-cli auth'",
		}
	}

	client := newXParserClient(cmd, credSrc, isFree)

	opts := &ParseOptions{
		PageRange:          parsePageRangeFlag,
		Password:           parsePassword,
		IncludeCharDetails: parseIncludeCharDetail,
	}

	if isFree {
		output.Status("Using free API")
	} else {
		output.Status("Using paid API (credentials: %s)", credSrc.Source)
	}

	if len(sources) == 1 {
		return runSingleParse(client, sources[0], opts)
	}
	return runBatchParse(client, sources, opts)
}

// ── single file/url ──

func runSingleParse(client *XParserClient, source string, opts *ParseOptions) error {
	output.Status("Parsing... %s", filepath.Base(source))
	start := time.Now()

	var resp *ParseResponse
	var err error

	if isURL(source) {
		resp, err = client.ParseURL(source, opts)
	} else {
		resp, err = client.ParseFile(source, opts)
	}
	if err != nil {
		return &exitError{code: exitcode.GeneralError, msg: err.Error()}
	}

	if resp.Code != 200 {
		return handleAPICodeError(resp.Code, resp.Message)
	}

	if !resp.HasResult() {
		return &exitError{code: exitcode.GeneralError, msg: "API returned success but no result data"}
	}

	elapsed := time.Since(start).Seconds()
	output.Status("Done in %.1fs (engine: %.0fms, pages: %d/%d)",
		elapsed, resp.GetDurationMs(), resp.GetSuccessCount(), resp.GetPageCount())

	return outputResult(resp, source)
}

// ── batch ──

func runBatchParse(client *XParserClient, sources []string, opts *ParseOptions) error {
	if err := os.MkdirAll(parseOutput, 0o755); err != nil {
		return &exitError{code: exitcode.GeneralError, msg: fmt.Sprintf("failed to create output directory: %v", err)}
	}

	total := len(sources)
	output.Status("Batch: %d files", total)

	succeeded := 0
	failed := 0
	start := time.Now()

	for i, source := range sources {
		output.Status("[%d/%d] Parsing... %s", i+1, total, filepath.Base(source))
		itemStart := time.Now()

		var resp *ParseResponse
		var err error

		if isURL(source) {
			resp, err = client.ParseURL(source, opts)
		} else {
			resp, err = client.ParseFile(source, opts)
		}

		if err != nil {
			output.Errorf("[%d/%d] %s - %v", i+1, total, filepath.Base(source), err)
			failed++
			continue
		}

		if resp.Code != 200 {
			info := exitcode.FromAPICode(resp.Code, resp.Message)
			output.Errorf("[%d/%d] %s - %s", i+1, total, filepath.Base(source), info.Message)
			if info.Suggestion != "" {
				output.Status("  Hint: %s", info.Suggestion)
			}
			failed++
			continue
		}

		if !resp.HasResult() {
			output.Errorf("[%d/%d] %s - API returned success but no result data", i+1, total, filepath.Base(source))
			failed++
			continue
		}

		saved, err := saveResult(resp, source, parseOutput)
		if err != nil {
			output.Errorf("[%d/%d] %s - save failed: %v", i+1, total, filepath.Base(source), err)
			failed++
			continue
		}

		output.Status("[%d/%d] Done: %s -> %s (%.1fs)",
			i+1, total, filepath.Base(source), saved, time.Since(itemStart).Seconds())
		succeeded++
	}

	elapsed := time.Since(start).Seconds()
	if failed > 0 {
		output.Status("Result: %d/%d succeeded, %d failed (%.1fs)", succeeded, total, failed, elapsed)
		return &exitError{code: exitcode.GeneralError, msg: fmt.Sprintf("batch completed with errors: %d/%d failed", failed, total)}
	}
	output.Status("Result: %d/%d succeeded (%.1fs)", succeeded, total, elapsed)
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
	saved, err := saveResult(resp, source, parseOutput)
	if err != nil {
		return &exitError{code: exitcode.GeneralError, msg: err.Error()}
	}
	output.Status("Saved: %s", saved)
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

// ── error handling ──

func handleAPICodeError(code int, message string) error {
	info := exitcode.FromAPICode(code, message)
	if info == nil {
		return nil
	}
	// Write structured error JSON to stderr
	output.Errorf("%s", info.JSON())
	return &exitError{code: exitcode.APIError, msg: info.Message}
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
