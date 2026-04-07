package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
)

// API base URLs.
const (
	paidAPIBaseURL = "https://api.textin.com"
	freeAPIBaseURL = "http://ht-pdf2md-sandbox.ai.intsig.net"
	parseAPIPath   = "/ai/service/v1/pdf_to_markdown"
)

// APIMode represents free vs paid API selection.
type APIMode string

const (
	APIModeAuto APIMode = ""     // auto: paid if key exists, else free
	APIModeFree APIMode = "free"
	APIModePaid APIMode = "paid"
)

// XParserClient wraps HTTP calls to the Textin xParser API.
type XParserClient struct {
	AppID      string
	SecretCode string
	BaseURL    string
	IsFreeAPI  bool
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
	Markdown         string          `json:"markdown"`
	Detail           json.RawMessage `json:"detail,omitempty"`
	Pages            json.RawMessage `json:"pages,omitempty"`
	Catalog          json.RawMessage `json:"catalog,omitempty"`
	TotalPageNumber  int             `json:"total_page_number,omitempty"`
	ValidPageNumber  int             `json:"valid_page_number,omitempty"`
	ExcelBase64      string          `json:"excel_base64,omitempty"`
	Elements         json.RawMessage `json:"elements,omitempty"`
}

// ParseOptions holds V1 parse parameters.
type ParseOptions struct {
	PageRange          string // e.g. "1-5" or "1-2,5-10"
	Password           string
	IncludeCharDetails bool
}

// buildQueryParams constructs query parameters with V1 defaults.
// Free and paid APIs share the same parameter schema; only the base URL differs.
func (o *ParseOptions) buildQueryParams() (url.Values, error) {
	q := url.Values{}

	// ── V1 defaults — always sent ──
	q.Set("apply_document_tree", "1") // include_hierarchy = true
	q.Set("get_image", "objects")     // include_inline_objects + include_image_data = true
	q.Set("table_flavor", "html")     // table_view = "html"
	q.Set("markdown_details", "1")    // include detail
	q.Set("page_details", "1")        // pages = true
	q.Set("catalog_details", "1")     // title_tree = true
	q.Set("apply_merge", "1")         // merge paragraphs/tables

	// ── Optional: include_char_details ──
	if o.IncludeCharDetails {
		q.Set("char_details", "1")
	}

	// ── Optional: password ──
	if o.Password != "" {
		q.Set("pdf_pwd", o.Password)
	}

	// ── Optional: page range ──
	if o.PageRange != "" {
		start, count, err := parsePageRange(o.PageRange)
		if err != nil {
			return nil, fmt.Errorf("invalid --page-range %q: %w", o.PageRange, err)
		}
		q.Set("page_start", fmt.Sprintf("%d", start))
		q.Set("page_count", fmt.Sprintf("%d", count))
	}

	return q, nil
}

// parsePageRange converts "1-5" or "1-2,5-10" to 0-based page_start and page_count.
// Page numbers in the range string are 1-based.
func parsePageRange(rangeStr string) (pageStart, pageCount int, err error) {
	minPage := math.MaxInt32
	maxPage := 0

	parts := strings.Split(rangeStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		bounds := strings.SplitN(part, "-", 2)
		if len(bounds) == 1 {
			p, e := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if e != nil || p < 1 {
				return 0, 0, fmt.Errorf("invalid page number: %s", bounds[0])
			}
			if p < minPage {
				minPage = p
			}
			if p > maxPage {
				maxPage = p
			}
		} else {
			start, e := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if e != nil || start < 1 {
				return 0, 0, fmt.Errorf("invalid range start: %s", bounds[0])
			}
			end, e := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if e != nil || end < start {
				return 0, 0, fmt.Errorf("invalid range end: %s", bounds[1])
			}
			if start < minPage {
				minPage = start
			}
			if end > maxPage {
				maxPage = end
			}
		}
	}

	if minPage == math.MaxInt32 {
		return 0, 0, fmt.Errorf("empty page range")
	}

	// Convert 1-based to 0-based
	return minPage - 1, maxPage - minPage + 1, nil
}

// resolveAPIMode determines whether to use free or paid API.
// Logic:
//   - --api free  → free
//   - --api paid  → paid (requires credentials)
//   - default     → paid if credentials exist, else free
func resolveAPIMode(mode APIMode, cred *config.CredentialSource) (isFree bool) {
	switch mode {
	case APIModeFree:
		return true
	case APIModePaid:
		return false
	default:
		// Auto: use paid if credentials are available
		return cred.AppID == "" || cred.SecretCode == ""
	}
}

// newXParserClient creates a client configured for free or paid API.
func newXParserClient(cmd *cobra.Command, cred *config.CredentialSource, isFree bool) *XParserClient {
	cfg, _ := config.Load()

	var baseURL string
	if isFree {
		baseURL = freeAPIBaseURL
	} else {
		baseURL = config.GetBaseURL(cmd, cfg)
	}

	httpClient := &http.Client{}
	if verboseFlag {
		httpClient = newVerboseHTTPClient()
	}

	return &XParserClient{
		AppID:      cred.AppID,
		SecretCode: cred.SecretCode,
		BaseURL:    baseURL,
		IsFreeAPI:  isFree,
		HTTPClient: httpClient,
	}
}

// ParseFile uploads a local file to the xParser API and returns the response.
func (c *XParserClient) ParseFile(filePath string, opts *ParseOptions) (*ParseResponse, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	apiURL, err := c.buildURL(opts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	c.setAuthHeaders(req)

	return c.doRequest(req)
}

// ParseURL sends a URL to the xParser API for remote file parsing.
func (c *XParserClient) ParseURL(fileURL string, opts *ParseOptions) (*ParseResponse, error) {
	apiURL, err := c.buildURL(opts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(fileURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	c.setAuthHeaders(req)

	return c.doRequest(req)
}

func (c *XParserClient) buildURL(opts *ParseOptions) (string, error) {
	apiURL := c.BaseURL + parseAPIPath
	q, err := opts.buildQueryParams()
	if err != nil {
		return "", err
	}
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}
	return apiURL, nil
}

func (c *XParserClient) setAuthHeaders(req *http.Request) {
	if !c.IsFreeAPI && c.AppID != "" && c.SecretCode != "" {
		req.Header.Set("x-ti-app-id", c.AppID)
		req.Header.Set("x-ti-secret-code", c.SecretCode)
	}
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

	var result ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
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
