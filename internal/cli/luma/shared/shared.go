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
	// LumaAPIBase is the base URL for Luma API
	LumaAPIBase = "https://api.lumalabs.ai/dream-machine/v1"
)

// Generation states
const (
	StateQueued    = "queued"
	StateDreaming  = "dreaming"
	StateCompleted = "completed"
	StateFailed    = "failed"
)

// GetLumaAPIKey returns the Luma API key
func GetLumaAPIKey() string {
	return config.GetAPIKey("LUMA_API_KEY")
}

// CreateRequest creates an HTTP request with Luma authentication headers
func CreateRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, LumaAPIBase+endpoint, body)
	if err != nil {
		return nil, err
	}

	apiKey := GetLumaAPIKey()
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

// Generation is the response from Luma API
type Generation struct {
	ID             string  `json:"id"`
	GenerationType string  `json:"generation_type,omitempty"`
	State          string  `json:"state"`
	FailureReason  string  `json:"failure_reason,omitempty"`
	CreatedAt      string  `json:"created_at,omitempty"`
	Model          string  `json:"model,omitempty"`
	Assets         *Assets `json:"assets,omitempty"`
}

// Assets contains the generated media URLs
type Assets struct {
	Video         string `json:"video,omitempty"`
	Image         string `json:"image,omitempty"`
	ProgressVideo string `json:"progress_video,omitempty"`
}

// ListResponse is the response for list generations
type ListResponse struct {
	HasMore     bool         `json:"has_more"`
	Count       int          `json:"count"`
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
	Generations []Generation `json:"generations"`
}

// HandleAPIError handles API error response
func HandleAPIError(cmd *cobra.Command, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Detail string `json:"detail"`
	}
	json.Unmarshal(body, &errResp)

	msg := errResp.Detail
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
		return common.WriteError(cmd, "server_error", "Luma API server error")
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
		return common.WriteError(cmd, "connection_error", "cannot connect to Luma API")
	}
	return common.WriteError(cmd, "network_error", err.Error())
}
