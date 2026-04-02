package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
	"github.com/textin/xparser-ecosystem/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	downloadOutput string
	downloadFrom   string
)

// downloadResponse is the JSON response from the image download API.
type downloadResponse struct {
	Code int `json:"code"`
	Data struct {
		Image string `json:"image"`
	} `json:"data"`
	Message string `json:"message,omitempty"`
}

var downloadCmd = &cobra.Command{
	Use:   "download [image_id ...] [--from result.json]",
	Short: "Download images from parse results",
	Long: `Download page images or sub-images from xParser API responses.

Two modes:

  1. By image ID:  ./xparser download <id1> <id2> -o ./images/
  2. From JSON:    ./xparser download --from result.json -o ./images/

With --from, the command reads a parse result JSON file and extracts all
image IDs from these locations:
  - metrics[].image_id          (page images)
  - result.pages[].image_id     (page images)
  - result.detail[].image_url   (sub-images)

Downloaded images are valid for 30 days from parsing.`,
	Example: `  ./xparser download abc123def456                       # download by ID
  ./xparser download abc123 def456 -o ./images/          # batch by IDs
  ./xparser download --from result.json -o ./images/     # extract IDs from JSON and download
  ./xparser download --from ./out/report.json            # from parse output JSON`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", ".", "Output path (file or directory)")
	downloadCmd.Flags().StringVar(&downloadFrom, "from", "", "Parse result JSON file to extract image IDs from")
}

// extractImageIDs reads a parse result JSON and extracts all image_id values.
func extractImageIDs(jsonPath string) ([]string, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read JSON file: %w", err)
	}

	// Use a flexible structure to handle the parse result JSON
	var result struct {
		Metrics []struct {
			ImageID string `json:"image_id"`
		} `json:"metrics"`
		Result struct {
			Pages []struct {
				ImageID string `json:"image_id"`
			} `json:"pages"`
			Detail []struct {
				ImageURL string `json:"image_url"`
			} `json:"detail"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	seen := make(map[string]bool)
	var ids []string

	// metrics[].image_id
	for _, m := range result.Metrics {
		id := strings.TrimSpace(m.ImageID)
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	// result.pages[].image_id
	for _, p := range result.Result.Pages {
		id := strings.TrimSpace(p.ImageID)
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	// result.detail[].image_url (sub-images)
	for _, d := range result.Result.Detail {
		url := strings.TrimSpace(d.ImageURL)
		if url != "" && !seen[url] {
			seen[url] = true
			ids = append(ids, url)
		}
	}

	return ids, nil
}

func runDownload(cmd *cobra.Command, args []string) error {
	// Determine image IDs: from --from JSON file or from args
	imageIDs := args
	if downloadFrom != "" {
		extracted, err := extractImageIDs(downloadFrom)
		if err != nil {
			return err
		}
		if len(extracted) == 0 {
			return fmt.Errorf("no image IDs found in %s", downloadFrom)
		}
		output.Status("Found %d image(s) in %s", len(extracted), downloadFrom)
		imageIDs = append(imageIDs, extracted...)
	}

	if len(imageIDs) == 0 {
		return fmt.Errorf("no image IDs specified. Provide IDs as arguments or use --from <result.json>")
	}

	credSrc, err := config.ResolveCredentials(cmd)
	if err != nil {
		return err
	}
	if credSrc.AppID == "" || credSrc.SecretCode == "" {
		return fmt.Errorf("no API credentials found. Run './xparser auth' to configure")
	}

	cfg, _ := config.Load()
	baseURL := config.GetBaseURL(cmd, cfg)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	if verboseFlag {
		httpClient = newVerboseHTTPClient()
		httpClient.Timeout = 60 * time.Second
	}

	if len(imageIDs) > 1 || isDirectory(downloadOutput) {
		if err := os.MkdirAll(downloadOutput, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	succeeded := 0
	for i, imageID := range imageIDs {
		outPath := downloadOutput
		if len(imageIDs) > 1 || isDirectory(downloadOutput) {
			dir := downloadOutput
			if !isDirectory(dir) {
				dir = filepath.Dir(dir)
			}
			os.MkdirAll(dir, 0o755)
			filename := filepath.Base(imageID)
			if !strings.HasSuffix(strings.ToLower(filename), ".jpg") && !strings.HasSuffix(strings.ToLower(filename), ".jpeg") && !strings.HasSuffix(strings.ToLower(filename), ".png") {
				filename += ".jpg"
			}
			outPath = filepath.Join(dir, filename)
		}

		output.Status("[%d/%d] Downloading %s...", i+1, len(imageIDs), imageID)

		apiURL := fmt.Sprintf("%s/ocr_image/download?image_id=%s", baseURL, url.QueryEscape(imageID))
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			output.Errorf("[%d/%d] %s - %v", i+1, len(imageIDs), imageID, err)
			continue
		}
		req.Header.Set("x-ti-app-id", credSrc.AppID)
		req.Header.Set("x-ti-secret-code", credSrc.SecretCode)

		resp, err := httpClient.Do(req)
		if err != nil {
			output.Errorf("[%d/%d] %s - request failed: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			output.Errorf("[%d/%d] %s - HTTP %d", i+1, len(imageIDs), imageID, resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			output.Errorf("[%d/%d] %s - read failed: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		// Parse JSON response and decode base64 image
		var dlResp downloadResponse
		if err := json.Unmarshal(body, &dlResp); err != nil {
			output.Errorf("[%d/%d] %s - invalid JSON response: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		if dlResp.Code != 200 {
			output.Errorf("[%d/%d] %s - API error %d: %s", i+1, len(imageIDs), imageID, dlResp.Code, dlResp.Message)
			continue
		}

		if dlResp.Data.Image == "" {
			output.Errorf("[%d/%d] %s - empty image data in response", i+1, len(imageIDs), imageID)
			continue
		}

		imageData, err := base64.StdEncoding.DecodeString(dlResp.Data.Image)
		if err != nil {
			output.Errorf("[%d/%d] %s - base64 decode failed: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			output.Errorf("[%d/%d] %s - mkdir failed: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		if err := os.WriteFile(outPath, imageData, 0o644); err != nil {
			output.Errorf("[%d/%d] %s - write failed: %v", i+1, len(imageIDs), imageID, err)
			continue
		}

		output.Status("[%d/%d] Saved: %s (%s)", i+1, len(imageIDs), outPath, humanSize(len(imageData)))
		succeeded++
	}

	if succeeded < len(imageIDs) {
		return fmt.Errorf("%d/%d downloads failed", len(imageIDs)-succeeded, len(imageIDs))
	}
	return nil
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func humanSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
}
