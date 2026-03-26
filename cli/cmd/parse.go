package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
	"github.com/textin/xparser-ecosystem/cli/internal/exitcode"
	"github.com/textin/xparser-ecosystem/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	parseOutput           string
	parseFormat           string
	parseParseMode        string
	parsePdfPwd           string
	parsePageStart        int
	parsePageCount        int
	parseDPI              int
	parseApplyDocTree     int
	parseTableFlavor      string
	parseGetImage         string
	parseImageOutputType  string
	parseParatextMode     string
	parseFormulaLevel     int
	parseUnderlineLevel   int
	parseApplyMerge       int
	parseApplyImageAnalysis int
	parseMarkdownDetails  int
	parsePageDetails      int
	parseRawOCR           int
	parseCharDetails      int
	parseCatalogDetails   int
	parseGetExcel         int
	parseCropDewarp       int
	parseRemoveWatermark  int
	parseApplyChart       int
	parseTimeout          int
	parseListFile         string
	parseStdinList        bool
	parseStdin            bool
	parseStdinName        string
)

var parseCmd = &cobra.Command{
	Use:   "parse <file-or-url> [...]",
	Short: "Parse documents to Markdown (Auth Required)",
	Long: `Parse documents using the Textin xParser API.

Supports: PDF, Images (png, jpg, bmp, tiff, webp), Doc(x), Ppt(x), Xls(x),
          HTML, MHTML, CSV, TXT, OFD, RTF

Features:
  - Layout detection, text recognition, table recognition
  - Formula recognition, heading hierarchy, image extraction
  - Multiple parse modes: auto, scan, lite, parse, vlm`,
	Example: `  xparser parse report.pdf                                  # markdown to stdout
  xparser parse report.pdf -o ./out/                        # save to directory
  xparser parse report.pdf -o report.md                     # save to specific file
  xparser parse report.pdf --parse-mode vlm                 # use VLM mode
  xparser parse report.pdf --table-flavor md                # markdown tables
  xparser parse report.pdf --get-image objects -o ./out/    # extract images
  xparser parse *.pdf -o ./results/                          # batch mode
  xparser parse --list files.txt -o ./results/               # batch from file list
  xparser parse https://example.com/doc.pdf                 # parse URL directly
  cat report.pdf | xparser parse --stdin -o report.md       # read from stdin`,
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	// Output
	parseCmd.Flags().StringVarP(&parseOutput, "output", "o", "", "Output path (file or dir); omit for stdout")
	parseCmd.Flags().StringVarP(&parseFormat, "format", "f", "md", "Output format: md, json (comma-separated)")

	// Parse mode
	parseCmd.Flags().StringVar(&parseParseMode, "parse-mode", "scan", "Parse mode: auto, scan, lite, parse, vlm")
	parseCmd.Flags().StringVar(&parsePdfPwd, "pdf-pwd", "", "PDF password for encrypted files")

	// Page range
	parseCmd.Flags().IntVar(&parsePageStart, "page-start", 0, "Start page for PDF (0-based)")
	parseCmd.Flags().IntVar(&parsePageCount, "page-count", 1000, "Number of pages to parse (max 1000)")

	// DPI
	parseCmd.Flags().IntVar(&parseDPI, "dpi", 144, "PDF coordinate DPI: 72, 144, or 216")

	// Document structure
	parseCmd.Flags().IntVar(&parseApplyDocTree, "apply-document-tree", 1, "Generate heading hierarchy: 0=no, 1=yes")
	parseCmd.Flags().StringVar(&parseTableFlavor, "table-flavor", "html", "Table format: md, html, none")
	parseCmd.Flags().StringVar(&parseParatextMode, "paratext-mode", "annotation", "Non-body text mode: none, annotation, body")
	parseCmd.Flags().IntVar(&parseApplyMerge, "apply-merge", 1, "Merge paragraphs/tables: 0=no, 1=yes")

	// Image handling
	parseCmd.Flags().StringVar(&parseGetImage, "get-image", "none", "Image return mode: none, page, objects, both")
	parseCmd.Flags().StringVar(&parseImageOutputType, "image-output-type", "default", "Image output: default (url), base64str")

	// Recognition features
	parseCmd.Flags().IntVar(&parseFormulaLevel, "formula-level", 0, "Formula recognition: 0=all, 1=display-only, 2=none")
	parseCmd.Flags().IntVar(&parseUnderlineLevel, "underline-level", 0, "Underline detection: 0=none, 1=empty, 2=all")
	parseCmd.Flags().IntVar(&parseApplyImageAnalysis, "apply-image-analysis", 0, "LLM image analysis: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseApplyChart, "apply-chart", 0, "Chart recognition: 0=no, 1=yes")

	// Output detail control
	parseCmd.Flags().IntVar(&parseMarkdownDetails, "markdown-details", 1, "Return detail field: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parsePageDetails, "page-details", 1, "Return pages field: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseRawOCR, "raw-ocr", 0, "Return raw OCR results: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseCharDetails, "char-details", 0, "Return char positions: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseCatalogDetails, "catalog-details", 0, "Return catalog info: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseGetExcel, "get-excel", 0, "Return excel base64: 0=no, 1=yes")

	// Image preprocessing
	parseCmd.Flags().IntVar(&parseCropDewarp, "crop-dewarp", 0, "Edge correction: 0=no, 1=yes")
	parseCmd.Flags().IntVar(&parseRemoveWatermark, "remove-watermark", 0, "Remove watermark: 0=no, 1=yes")

	// Timeout & input modes
	parseCmd.Flags().IntVar(&parseTimeout, "timeout", 600, "Timeout in seconds")
	parseCmd.Flags().StringVar(&parseListFile, "list", "", "Read input list from file (one path per line)")
	parseCmd.Flags().BoolVar(&parseStdinList, "stdin-list", false, "Read input list from stdin")
	parseCmd.Flags().BoolVar(&parseStdin, "stdin", false, "Read file content from stdin")
	parseCmd.Flags().StringVar(&parseStdinName, "stdin-name", "stdin.pdf", "Filename hint for stdin mode")
}

func runParse(cmd *cobra.Command, args []string) error {
	sources, err := collectSources(args, parseListFile, parseStdinList)
	if err != nil {
		return err
	}

	if len(sources) == 0 && !parseStdin {
		return fmt.Errorf("no input files specified. Provide files as arguments, use --list, or --stdin")
	}

	formats := parseFormats(parseFormat)

	if err := validateOutputMode(parseOutput, formats, sources, parseStdin); err != nil {
		return err
	}

	credSrc, err := config.ResolveCredentials(cmd)
	if err != nil {
		return err
	}
	if credSrc.AppID == "" || credSrc.SecretCode == "" {
		return fmt.Errorf("no API credentials found. Run 'xparser auth' to configure your credentials")
	}

	client := newXParserClient(cmd, credSrc)
	opts := buildParseOpts(cmd)

	if parseStdin {
		return runStdinParse(client, opts)
	}
	if len(sources) == 1 {
		return runSingleParse(client, sources[0], formats, opts)
	}
	return runBatchParse(client, sources, formats, opts)
}

// ── single file/url ──

func runSingleParse(client *XParserClient, source string, formats []string, opts *ParseOptions) error {
	client.HTTPClient.Timeout = time.Duration(parseTimeout) * time.Second

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
		return handleAPIError(err)
	}

	if resp.Code != 200 {
		return handleAPICodeError(resp.Code, resp.Message)
	}

	elapsed := time.Since(start).Seconds()
	output.Status("Done in %.1fs (engine: %dms, pages: %d/%d)",
		elapsed, resp.Duration,
		resp.Result.ValidPageNumber, resp.Result.TotalPageNumber)

	return outputParseResult(resp, source, formats)
}

// ── stdin ──

func runStdinParse(client *XParserClient, opts *ParseOptions) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data received from stdin")
	}

	tmpDir, err := os.MkdirTemp("", "xparser-stdin-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := filepath.Join(tmpDir, parseStdinName)
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	formats := parseFormats(parseFormat)
	return runSingleParse(client, tmpPath, formats, opts)
}

// ── batch ──

func runBatchParse(client *XParserClient, sources, formats []string, opts *ParseOptions) error {
	client.HTTPClient.Timeout = time.Duration(parseTimeout) * time.Second

	if err := os.MkdirAll(parseOutput, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
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
			if info.Hint != "" {
				output.Status("  Hint: %s", info.Hint)
			}
			failed++
			continue
		}

		saved, err := saveParseResult(resp, source, formats, parseOutput)
		if err != nil {
			output.Errorf("[%d/%d] %s - save failed: %v", i+1, total, filepath.Base(source), err)
			failed++
			continue
		}

		output.Status("[%d/%d] Done: %s -> %s (%.1fs)",
			i+1, total, filepath.Base(source),
			strings.Join(saved, ", "), time.Since(itemStart).Seconds())
		succeeded++
	}

	elapsed := time.Since(start).Seconds()
	if failed > 0 {
		output.Status("Result: %d/%d succeeded, %d failed (%.1fs)", succeeded, total, failed, elapsed)
		return fmt.Errorf("batch completed with errors: %d/%d failed", failed, total)
	}
	output.Status("Result: %d/%d succeeded (%.1fs)", succeeded, total, elapsed)
	return nil
}

// ── output ──

func outputParseResult(resp *ParseResponse, source string, formats []string) error {
	if parseOutput == "" {
		// stdout mode
		f := formats[0]
		switch f {
		case "md":
			fmt.Print(resp.Result.Markdown)
		case "json":
			data, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(data))
		}
		return nil
	}

	// file mode
	saved, err := saveParseResult(resp, source, formats, parseOutput)
	if err != nil {
		return err
	}
	output.Status("Saved: %s", strings.Join(saved, ", "))
	return nil
}

func saveParseResult(resp *ParseResponse, source string, formats []string, outputDir string) ([]string, error) {
	dir, base := resolveOutputTarget(outputDir, source, formats)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var saved []string
	for _, f := range formats {
		switch f {
		case "md":
			p := filepath.Join(dir, base+".md")
			if err := os.WriteFile(p, []byte(resp.Result.Markdown), 0o644); err != nil {
				return saved, fmt.Errorf("failed to save markdown: %w", err)
			}
			saved = append(saved, p)
		case "json":
			p := filepath.Join(dir, base+".json")
			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return saved, fmt.Errorf("failed to marshal json: %w", err)
			}
			if err := os.WriteFile(p, data, 0o644); err != nil {
				return saved, fmt.Errorf("failed to save json: %w", err)
			}
			saved = append(saved, p)
		}
	}

	// Save excel if requested and available
	if parseGetExcel == 1 && resp.Result.ExcelBase64 != "" {
		p := filepath.Join(dir, base+".xlsx")
		if err := SaveExcel(resp.Result.ExcelBase64, p); err != nil {
			output.Status("Warning: failed to save excel: %v", err)
		} else {
			saved = append(saved, p)
		}
	}

	return saved, nil
}

// ── options builder ──

func buildParseOpts(cmd *cobra.Command) *ParseOptions {
	opts := NewParseOptions()

	flagMap := map[string]*struct {
		strVal  *string
		intVal  *int
	}{
		"parse-mode":            {strVal: &opts.ParseMode},
		"table-flavor":          {strVal: &opts.TableFlavor},
		"get-image":             {strVal: &opts.GetImage},
		"image-output-type":     {strVal: &opts.ImageOutputType},
		"paratext-mode":         {strVal: &opts.ParatextMode},
	}

	intFlagMap := map[string]*int{
		"page-start":            &opts.PageStart,
		"page-count":            &opts.PageCount,
		"dpi":                   &opts.DPI,
		"apply-document-tree":   &opts.ApplyDocumentTree,
		"formula-level":         &opts.FormulaLevel,
		"underline-level":       &opts.UnderlineLevel,
		"apply-merge":           &opts.ApplyMerge,
		"apply-image-analysis":  &opts.ApplyImageAnalysis,
		"markdown-details":      &opts.MarkdownDetails,
		"page-details":          &opts.PageDetails,
		"raw-ocr":               &opts.RawOCR,
		"char-details":          &opts.CharDetails,
		"catalog-details":       &opts.CatalogDetails,
		"get-excel":             &opts.GetExcel,
		"crop-dewarp":           &opts.CropDewarp,
		"remove-watermark":      &opts.RemoveWatermark,
		"apply-chart":           &opts.ApplyChart,
	}

	for name, f := range flagMap {
		if cmd.Flags().Changed(name) {
			opts.SetChanged(name)
			if f.strVal != nil {
				val, _ := cmd.Flags().GetString(name)
				*f.strVal = val
			}
		}
	}

	for name, ptr := range intFlagMap {
		if cmd.Flags().Changed(name) {
			opts.SetChanged(name)
			val, _ := cmd.Flags().GetInt(name)
			*ptr = val
		}
	}

	opts.PdfPwd = parsePdfPwd

	return opts
}

// ── validation ──

func validateOutputMode(outputPath string, formats []string, sources []string, isStdin bool) error {
	count := len(sources)
	if isStdin {
		count = 1
	}

	if outputPath != "" {
		if count > 1 && outputPathLooksLikeFile(outputPath, formats) {
			return fmt.Errorf("batch mode requires -o to specify output directory, not a file path")
		}
		if len(formats) > 1 && outputPathLooksLikeFile(outputPath, formats) {
			return fmt.Errorf("multiple formats require -o to specify an output directory, not a file path")
		}
		return nil
	}

	if count > 1 {
		return fmt.Errorf("batch mode requires -o to specify output directory")
	}
	if len(formats) > 1 {
		return fmt.Errorf("multiple formats cannot output to stdout, use -o to save to file")
	}
	return nil
}

// ── error handling ──

func handleAPIError(err error) error {
	output.Errorf("%s", err.Error())
	os.Exit(exitcode.GeneralError)
	return nil
}

func handleAPICodeError(code int, message string) error {
	info := exitcode.FromAPICode(code, message)
	if info == nil {
		return nil
	}
	output.Errorf("%s", info.Message)
	if info.Hint != "" {
		output.Status("Hint: %s", info.Hint)
	}
	os.Exit(info.Code)
	return nil
}

// ── shared helpers ──

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func collectSources(args []string, listFile string, stdinList bool) ([]string, error) {
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

	if stdinList {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				sources = append(sources, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
	}

	return sources, nil
}

func parseFormats(raw string) []string {
	if raw == "" {
		return []string{"md"}
	}
	var formats []string
	for _, f := range strings.Split(raw, ",") {
		f = strings.TrimSpace(strings.ToLower(f))
		if f != "" {
			formats = append(formats, f)
		}
	}
	if len(formats) == 0 {
		return []string{"md"}
	}
	return formats
}

func baseNameNoExt(source string) string {
	base := filepath.Base(source)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

func resolveOutputTarget(outputPath, source string, formats []string) (dir, base string) {
	dir = outputPath
	base = baseNameNoExt(source)

	info, err := os.Stat(outputPath)
	if err == nil && !info.IsDir() {
		return filepath.Dir(outputPath), baseNameNoExt(outputPath)
	}

	if outputPathLooksLikeFile(outputPath, formats) {
		return filepath.Dir(outputPath), baseNameNoExt(outputPath)
	}

	return dir, base
}

func outputPathLooksLikeFile(outputPath string, formats []string) bool {
	if outputPath == "" || hasTrailingPathSeparator(outputPath) {
		return false
	}

	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == "" {
		return false
	}

	for _, candidate := range outputExtensions(formats) {
		if ext == candidate {
			return true
		}
	}
	return false
}

func outputExtensions(formats []string) []string {
	seen := make(map[string]struct{})
	var exts []string
	for _, format := range formats {
		for _, ext := range formatExtensions(format) {
			if _, ok := seen[ext]; ok {
				continue
			}
			seen[ext] = struct{}{}
			exts = append(exts, ext)
		}
	}
	return exts
}

func formatExtensions(format string) []string {
	switch format {
	case "md":
		return []string{".md", ".markdown"}
	case "json":
		return []string{".json"}
	default:
		return nil
	}
}

func hasTrailingPathSeparator(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\")
}
