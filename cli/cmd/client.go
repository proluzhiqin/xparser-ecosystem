package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
	"github.com/spf13/cobra"
)

// XParserClient wraps HTTP calls to the Textin xParser API.
type XParserClient struct {
	AppID      string
	SecretCode string
	BaseURL    string
	HTTPClient *http.Client
}

// ParseResponse is the top-level JSON response from the xParser API.
type ParseResponse struct {
	Code     int             `json:"code"`
	Message  string          `json:"message"`
	Result   *ParseResult    `json:"result,omitempty"`
	Version  string          `json:"version,omitempty"`
	Duration int             `json:"duration,omitempty"`
	Metrics  json.RawMessage `json:"metrics,omitempty"`
}

// ParseResult holds the parsed output from the API.
type ParseResult struct {
	Markdown         string            `json:"markdown"`
	Detail           json.RawMessage   `json:"detail,omitempty"`
	Pages            json.RawMessage   `json:"pages,omitempty"`
	Catalog          json.RawMessage   `json:"catalog,omitempty"`
	TotalPageNumber  int               `json:"total_page_number,omitempty"`
	ValidPageNumber  int               `json:"valid_page_number,omitempty"`
	ExcelBase64      string            `json:"excel_base64,omitempty"`
	Elements         json.RawMessage   `json:"elements,omitempty"`
}

// ParseOptions contains the query parameters for the parse API.
type ParseOptions struct {
	ParseMode         string
	PdfPwd            string
	PageStart         int
	PageCount         int
	DPI               int
	ApplyDocumentTree int
	TableFlavor       string
	GetImage          string
	ImageOutputType   string
	ParatextMode      string
	FormulaLevel      int
	UnderlineLevel    int
	ApplyMerge        int
	ApplyImageAnalysis int
	MarkdownDetails   int
	PageDetails       int
	RawOCR            int
	CharDetails       int
	CatalogDetails    int
	GetExcel          int
	CropDewarp        int
	RemoveWatermark   int
	ApplyChart        int

	// Track which flags were explicitly set
	changed map[string]bool
}

// NewParseOptions returns defaults matching the API defaults.
func NewParseOptions() *ParseOptions {
	return &ParseOptions{
		ParseMode:         "scan",
		PageStart:         0,
		PageCount:         1000,
		DPI:               144,
		ApplyDocumentTree: 1,
		TableFlavor:       "html",
		GetImage:          "none",
		ImageOutputType:   "default",
		ParatextMode:      "annotation",
		FormulaLevel:      0,
		UnderlineLevel:    0,
		ApplyMerge:        1,
		ApplyImageAnalysis: 0,
		MarkdownDetails:   1,
		PageDetails:       1,
		RawOCR:            0,
		CharDetails:       0,
		CatalogDetails:    0,
		GetExcel:          0,
		CropDewarp:        0,
		RemoveWatermark:   0,
		ApplyChart:        0,
		changed:           make(map[string]bool),
	}
}

// SetChanged marks a flag as explicitly set by the user.
func (o *ParseOptions) SetChanged(name string) {
	if o.changed == nil {
		o.changed = make(map[string]bool)
	}
	o.changed[name] = true
}

// IsChanged returns whether a flag was explicitly set.
func (o *ParseOptions) IsChanged(name string) bool {
	return o.changed[name]
}

// buildQueryParams constructs query parameters, only including non-default or explicitly-changed values.
func (o *ParseOptions) buildQueryParams() url.Values {
	q := url.Values{}

	if o.IsChanged("parse-mode") {
		q.Set("parse_mode", o.ParseMode)
	}
	if o.PdfPwd != "" {
		q.Set("pdf_pwd", o.PdfPwd)
	}
	if o.IsChanged("page-start") {
		q.Set("page_start", fmt.Sprintf("%d", o.PageStart))
	}
	if o.IsChanged("page-count") {
		q.Set("page_count", fmt.Sprintf("%d", o.PageCount))
	}
	if o.IsChanged("dpi") {
		q.Set("dpi", fmt.Sprintf("%d", o.DPI))
	}
	if o.IsChanged("apply-document-tree") {
		q.Set("apply_document_tree", fmt.Sprintf("%d", o.ApplyDocumentTree))
	}
	if o.IsChanged("table-flavor") {
		q.Set("table_flavor", o.TableFlavor)
	}
	if o.IsChanged("get-image") {
		q.Set("get_image", o.GetImage)
	}
	if o.IsChanged("image-output-type") {
		q.Set("image_output_type", o.ImageOutputType)
	}
	if o.IsChanged("paratext-mode") {
		q.Set("paratext_mode", o.ParatextMode)
	}
	if o.IsChanged("formula-level") {
		q.Set("formula_level", fmt.Sprintf("%d", o.FormulaLevel))
	}
	if o.IsChanged("underline-level") {
		q.Set("underline_level", fmt.Sprintf("%d", o.UnderlineLevel))
	}
	if o.IsChanged("apply-merge") {
		q.Set("apply_merge", fmt.Sprintf("%d", o.ApplyMerge))
	}
	if o.IsChanged("apply-image-analysis") {
		q.Set("apply_image_analysis", fmt.Sprintf("%d", o.ApplyImageAnalysis))
	}
	if o.IsChanged("markdown-details") {
		q.Set("markdown_details", fmt.Sprintf("%d", o.MarkdownDetails))
	}
	if o.IsChanged("page-details") {
		q.Set("page_details", fmt.Sprintf("%d", o.PageDetails))
	}
	if o.IsChanged("raw-ocr") {
		q.Set("raw_ocr", fmt.Sprintf("%d", o.RawOCR))
	}
	if o.IsChanged("char-details") {
		q.Set("char_details", fmt.Sprintf("%d", o.CharDetails))
	}
	if o.IsChanged("catalog-details") {
		q.Set("catalog_details", fmt.Sprintf("%d", o.CatalogDetails))
	}
	if o.IsChanged("get-excel") {
		q.Set("get_excel", fmt.Sprintf("%d", o.GetExcel))
	}
	if o.IsChanged("crop-dewarp") {
		q.Set("crop_dewarp", fmt.Sprintf("%d", o.CropDewarp))
	}
	if o.IsChanged("remove-watermark") {
		q.Set("remove_watermark", fmt.Sprintf("%d", o.RemoveWatermark))
	}
	if o.IsChanged("apply-chart") {
		q.Set("apply_chart", fmt.Sprintf("%d", o.ApplyChart))
	}

	return q
}

// newXParserClient creates a client with credentials and global flags applied.
func newXParserClient(cmd *cobra.Command, cred *config.CredentialSource) *XParserClient {
	cfg, _ := config.Load()
	baseURL := config.GetBaseURL(cmd, cfg)

	httpClient := &http.Client{}
	if verboseFlag {
		httpClient = newVerboseHTTPClient()
	}

	return &XParserClient{
		AppID:      cred.AppID,
		SecretCode: cred.SecretCode,
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	}
}

// ParseFile uploads a local file to the xParser API and returns the response.
func (c *XParserClient) ParseFile(filePath string, opts *ParseOptions) (*ParseResponse, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	apiURL := c.BaseURL + "/ai/service/v1/pdf_to_markdown"
	q := opts.buildQueryParams()
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-ti-app-id", c.AppID)
	req.Header.Set("x-ti-secret-code", c.SecretCode)

	return c.doRequest(req)
}

// ParseURL sends a URL to the xParser API for remote file parsing.
func (c *XParserClient) ParseURL(fileURL string, opts *ParseOptions) (*ParseResponse, error) {
	apiURL := c.BaseURL + "/ai/service/v1/pdf_to_markdown"
	q := opts.buildQueryParams()
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(fileURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-ti-app-id", c.AppID)
	req.Header.Set("x-ti-secret-code", c.SecretCode)

	return c.doRequest(req)
}

func (c *XParserClient) doRequest(req *http.Request) (*ParseResponse, error) {

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return &result, nil
}

// SaveExcel decodes and saves the excel_base64 field to a file.
func SaveExcel(base64Str string, path string) error {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return fmt.Errorf("failed to decode excel base64: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
