package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/textin/xparser-ecosystem/cli/internal/config"
)

var authShow bool

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure API credentials for document parsing",
	Long: `Authenticate with your Textin xParser API credentials (app-id and secret-code).
Get your credentials at https://www.textin.com/console/dashboard/setting`,
	Example: `  xparse-cli auth              # Interactive credential setup
  xparse-cli auth --show       # Show current credential source

  # For automation, set environment variables:
  export XPARSE_APP_ID=your_app_id
  export XPARSE_SECRET_CODE=your_secret_code`,
	Args: cobra.NoArgs,
	RunE: runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().BoolVar(&authShow, "show", false, "Show current credential source")
}

func runAuth(cmd *cobra.Command, args []string) error {
	if authShow {
		return runAuthShow()
	}
	return runAuthSetup()
}

func runAuthShow() error {
	credSrc, err := config.ResolveCredentials(nil)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if credSrc.AppID == "" {
		fmt.Println("No credentials configured.")
		fmt.Println("Run 'xparse-cli auth' to set up your API credentials.")
		return nil
	}

	fmt.Printf("Credential source: %s\n", credSrc.Source)
	fmt.Printf("App ID:      %s\n", maskToken(credSrc.AppID))
	fmt.Printf("Secret Code: %s\n", maskToken(credSrc.SecretCode))

	cfg, err := config.Load()
	if err == nil && cfg.BaseURL != "" {
		fmt.Printf("Base URL:    %s\n", cfg.BaseURL)
	}
	return nil
}

func runAuthSetup() error {
	fmt.Println("Textin xParser API Credential Setup")
	fmt.Println("Get your credentials from: https://www.textin.com/console/dashboard/setting")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	existing, _ := config.ResolveCredentials(nil)
	if existing.AppID != "" {
		fmt.Printf("Current credential source: %s\n", existing.Source)
		fmt.Printf("Current App ID: %s\n", maskToken(existing.AppID))
		fmt.Println()
	}

	// Read App ID
	if existing.AppID != "" {
		fmt.Print("Enter new App ID (or press Enter to keep current): ")
	} else {
		fmt.Print("Enter your App ID (x-ti-app-id): ")
	}
	appID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	appID = strings.TrimSpace(appID)

	if appID == "" && existing.AppID != "" {
		appID = existing.AppID
	}
	if appID == "" {
		return fmt.Errorf("app-id is required")
	}

	// Read Secret Code
	if existing.SecretCode != "" {
		fmt.Print("Enter new Secret Code (or press Enter to keep current): ")
	} else {
		fmt.Print("Enter your Secret Code (x-ti-secret-code): ")
	}
	secretCode, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	secretCode = strings.TrimSpace(secretCode)

	if secretCode == "" && existing.SecretCode != "" {
		secretCode = existing.SecretCode
	}
	if secretCode == "" {
		return fmt.Errorf("secret-code is required")
	}

	if err := config.SetCredentials(appID, secretCode); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Credentials saved to ~/.xparse-cli/config.yaml")
	return nil
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
