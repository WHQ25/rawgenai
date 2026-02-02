package shared

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

const (
	// RunwayAPIBase is the base URL for Runway API
	RunwayAPIBase = "https://api.dev.runwayml.com"
	// RunwayAPIVersion is the required API version header
	RunwayAPIVersion = "2024-11-06"
)

// Task status constants
const (
	StatusPending   = "PENDING"
	StatusThrottled = "THROTTLED"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
)

// GetRunwayAPIKey returns the Runway API key
func GetRunwayAPIKey() string {
	return config.GetAPIKey("RUNWAY_API_KEY")
}

// CreateRequest creates an HTTP request with Runway authentication headers
func CreateRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, RunwayAPIBase+endpoint, body)
	if err != nil {
		return nil, err
	}

	apiKey := GetRunwayAPIKey()
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Runway-Version", RunwayAPIVersion)
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

// ResolveMediaURI resolves local file to data URI or returns URL as-is
func ResolveMediaURI(input string, mediaType string) (string, error) {
	if IsURL(input) {
		return input, nil
	}
	return EncodeToDataURI(input, mediaType)
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
	case "video":
		switch ext {
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		case ".mov":
			return "video/quicktime"
		default:
			return "video/mp4"
		}
	case "audio":
		switch ext {
		case ".mp3":
			return "audio/mpeg"
		case ".wav":
			return "audio/wav"
		case ".ogg":
			return "audio/ogg"
		case ".m4a":
			return "audio/mp4"
		default:
			return "audio/mpeg"
		}
	}
	return "application/octet-stream"
}

// GetPrompt gets prompt from args, file, or stdin
func GetPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// From positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(args[0])
		if text != "" {
			return text, nil
		}
	}

	// From file
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot read file: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("file is empty")
		}
		return text, nil
	}

	// From stdin
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stat, _ := f.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return "", fmt.Errorf("no prompt provided")
			}
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("cannot read stdin: %s", err.Error())
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", fmt.Errorf("stdin is empty")
		}
		return text, nil
	}

	return "", fmt.Errorf("no prompt provided")
}

// TaskResponse is the response from task creation
type TaskResponse struct {
	ID string `json:"id"`
}

// TaskStatus is the response from task status query
type TaskStatus struct {
	ID        string   `json:"id"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"createdAt"`
	Output    []string `json:"output,omitempty"`
	Failure   string   `json:"failure,omitempty"`
}

// HandleAPIError handles API error response
func HandleAPIError(cmd *cobra.Command, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Error string `json:"error"`
	}
	json.Unmarshal(body, &errResp)

	msg := errResp.Error
	if msg == "" {
		msg = string(body)
	}

	switch resp.StatusCode {
	case 400:
		return common.WriteError(cmd, "invalid_request", msg)
	case 401:
		return common.WriteError(cmd, "invalid_api_key", "API key is invalid")
	case 403:
		return common.WriteError(cmd, "forbidden", msg)
	case 404:
		return common.WriteError(cmd, "not_found", msg)
	case 429:
		return common.WriteError(cmd, "rate_limit", "Too many requests")
	case 500, 502, 503:
		return common.WriteError(cmd, "server_error", "Runway API server error")
	default:
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API error: %d - %s", resp.StatusCode, msg))
	}
}

// HandleHTTPError handles HTTP connection errors
func HandleHTTPError(cmd *cobra.Command, err error) error {
	errStr := err.Error()
	if strings.Contains(errStr, "timeout") {
		return common.WriteError(cmd, "timeout", "request timed out")
	}
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "refused") {
		return common.WriteError(cmd, "connection_error", "cannot connect to Runway API")
	}
	return common.WriteError(cmd, "network_error", err.Error())
}
