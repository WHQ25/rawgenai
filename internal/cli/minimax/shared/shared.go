package shared

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/config"
)

const (
	// MinimaxAPIBase is the base URL for MiniMax API
	MinimaxAPIBase = "https://api.minimax.io"
)

// GetMinimaxAPIKey returns the MiniMax API key
func GetMinimaxAPIKey() string {
	return config.GetAPIKey("MINIMAX_API_KEY")
}

// CreateRequest creates an HTTP request with MiniMax authentication headers
func CreateRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, MinimaxAPIBase+endpoint, body)
	if err != nil {
		return nil, err
	}

	apiKey := GetMinimaxAPIKey()
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// DoRequest executes HTTP request and returns response
func DoRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	return client.Do(req)
}

// DoRequestWithTimeout executes HTTP request with custom timeout
func DoRequestWithTimeout(req *http.Request, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	return client.Do(req)
}

// IsURL checks if string is a URL
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ResolveImageURL resolves local file to data URI or returns URL as-is
func ResolveImageURL(input string) (string, error) {
	if IsURL(input) {
		return input, nil
	}
	return EncodeToDataURI(input, "image")
}

// EncodeToDataURI encodes local file to data URI
func EncodeToDataURI(path string, mediaType string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	mimeType := DetectMIMEType(path, mediaType)
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// DetectMIMEType detects MIME type from file extension
func DetectMIMEType(path string, mediaType string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch mediaType {
	case "image":
		switch ext {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		default:
			return "image/jpeg"
		}
	}
	return "application/octet-stream"
}
