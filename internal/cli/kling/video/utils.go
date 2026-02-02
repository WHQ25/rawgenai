package video

import (
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
