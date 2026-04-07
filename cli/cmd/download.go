package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/textin/xparser-ecosystem/cli/internal/output"
)

var (
	downloadOutput string
	downloadFrom   string
)

var downloadCmd = &cobra.Command{
	Use:   "download [url ...] [--from result.json]",
	Short: "Download images from parse results",
	Long: `Download element images from xParser parse results.

Two modes:

  1. By URL:      xparse-cli download <url1> <url2> -o ./images/
  2. From JSON:   xparse-cli download --from result.json -o ./images/

With --from, the command reads a parse result JSON file and downloads all
images found in data.elements[].image_data.image_url.

Image URLs are publicly accessible and downloaded directly via HTTP GET.`,
	Example: `  xparse-cli download --from result.json -o ./images/
  xparse-cli download https://web-api.textin.com/ocr_image/external/abc123.jpg -o ./images/`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", ".", "Output path (file or directory)")
	downloadCmd.Flags().StringVar(&downloadFrom, "from", "", "Parse result JSON file to extract image URLs from")
}

// extractImageURLs reads a parse result JSON and extracts all image URLs.
func extractImageURLs(jsonPath string) ([]string, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read JSON file: %w", err)
	}

	var result struct {
		Data struct {
			Elements []struct {
				ImageData struct {
					ImageURL string `json:"image_url"`
				} `json:"image_data"`
			} `json:"elements"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	seen := make(map[string]bool)
	var urls []string

	for _, el := range result.Data.Elements {
		u := strings.TrimSpace(el.ImageData.ImageURL)
		if u != "" && !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}

	return urls, nil
}

func runDownload(cmd *cobra.Command, args []string) error {
	var imageURLs []string
	imageURLs = append(imageURLs, args...)

	if downloadFrom != "" {
		extracted, err := extractImageURLs(downloadFrom)
		if err != nil {
			return err
		}
		if len(extracted) == 0 {
			return fmt.Errorf("no image URLs found in %s", downloadFrom)
		}
		output.Status("Found %d image(s) in %s", len(extracted), downloadFrom)
		imageURLs = append(imageURLs, extracted...)
	}

	if len(imageURLs) == 0 {
		return fmt.Errorf("no image URLs specified. Provide URLs as arguments or use --from <result.json>")
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	if verboseFlag {
		httpClient = newVerboseHTTPClient()
		httpClient.Timeout = 60 * time.Second
	}

	if len(imageURLs) > 1 || isDirectory(downloadOutput) {
		if err := os.MkdirAll(downloadOutput, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	succeeded := 0
	isBatchOrDir := len(imageURLs) > 1 || isDirectory(downloadOutput)

	for i, imageURL := range imageURLs {
		// Derive output filename from URL
		outPath := downloadOutput
		if isBatchOrDir {
			filename := filepath.Base(imageURL)
			// Strip query string from filename
			if idx := strings.IndexByte(filename, '?'); idx != -1 {
				filename = filename[:idx]
			}
			if filename == "" || filename == "." {
				filename = fmt.Sprintf("image_%d.jpg", i+1)
			}
			// Ensure file has an image extension
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" && ext != ".bmp" && ext != ".tiff" {
				filename += ".jpg"
			}
			outPath = filepath.Join(downloadOutput, filename)
		}

		output.Status("[%d/%d] Downloading %s...", i+1, len(imageURLs), filepath.Base(imageURL))

		resp, err := httpClient.Get(imageURL)
		if err != nil {
			output.Errorf("[%d/%d] %s - request failed: %v", i+1, len(imageURLs), filepath.Base(imageURL), err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			output.Errorf("[%d/%d] %s - HTTP %d", i+1, len(imageURLs), filepath.Base(imageURL), resp.StatusCode)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			resp.Body.Close()
			output.Errorf("[%d/%d] %s - mkdir failed: %v", i+1, len(imageURLs), filepath.Base(imageURL), err)
			continue
		}

		f, err := os.Create(outPath)
		if err != nil {
			resp.Body.Close()
			output.Errorf("[%d/%d] %s - create file failed: %v", i+1, len(imageURLs), filepath.Base(imageURL), err)
			continue
		}

		n, err := io.Copy(f, resp.Body)
		resp.Body.Close()
		f.Close()
		if err != nil {
			output.Errorf("[%d/%d] %s - write failed: %v", i+1, len(imageURLs), filepath.Base(imageURL), err)
			continue
		}

		output.Status("[%d/%d] Saved: %s (%s)", i+1, len(imageURLs), outPath, humanSize(int(n)))
		succeeded++
	}

	if succeeded < len(imageURLs) {
		return fmt.Errorf("%d/%d downloads failed", len(imageURLs)-succeeded, len(imageURLs))
	}
	output.Status("Downloaded %d image(s)", succeeded)
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
