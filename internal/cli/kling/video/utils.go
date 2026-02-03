package video

import (
	"io"

	"github.com/spf13/cobra"
)

// Exported functions for use by other kling subpackages (e.g., voice)

// GetKlingAPIBase returns the Kling API base URL.
func GetKlingAPIBase() string {
	return getKlingAPIBase()
}

// GenerateJWT generates a JWT token for Kling API authentication.
func GenerateJWT(accessKey, secretKey string) (string, error) {
	return generateJWT(accessKey, secretKey)
}

// HandleAPIError handles network/connection errors and returns appropriate CLI error.
func HandleAPIError(cmd *cobra.Command, err error) error {
	return handleAPIError(cmd, err)
}

// HandleKlingError handles Kling API error codes and returns appropriate CLI error.
func HandleKlingError(cmd *cobra.Command, code int, message string) error {
	return handleKlingError(cmd, code, message)
}

// GetPrompt resolves prompt text from args, file, or stdin.
func GetPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	return getPrompt(args, filePath, stdin)
}

// ResolveImageURL returns the image URL or base64 string for API request.
func ResolveImageURL(input string) (string, error) {
	return resolveImageURL(input)
}
