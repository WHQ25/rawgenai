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

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
	aiart "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/aiart/v20221229"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vclm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vclm/v20240523"
)

const DefaultRegion = "ap-guangzhou"

// GetCredentials returns Tencent Cloud secret ID and key from config/env.
func GetCredentials() (secretID, secretKey string) {
	secretID = config.GetAPIKey("TENCENT_SECRET_ID")
	secretKey = config.GetAPIKey("TENCENT_SECRET_KEY")
	return
}

// CheckCredentials validates credentials and returns an error via WriteError if missing.
func CheckCredentials(cmd *cobra.Command) (string, string, error) {
	secretID, secretKey := GetCredentials()
	if secretID == "" || secretKey == "" {
		return "", "", common.WriteError(cmd, "missing_api_key",
			config.GetMissingKeyMessage("TENCENT_SECRET_ID")+" and "+config.GetMissingKeyMessage("TENCENT_SECRET_KEY"))
	}
	return secretID, secretKey, nil
}

// NewAiartClient creates a Tencent Cloud aiart SDK client.
func NewAiartClient(secretID, secretKey, region string) (*aiart.Client, error) {
	credential := tccommon.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	return aiart.NewClient(credential, region, cpf)
}

// NewVclmClient creates a Tencent Cloud vclm SDK client.
func NewVclmClient(secretID, secretKey, region string) (*vclm.Client, error) {
	credential := tccommon.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	return vclm.NewClient(credential, region, cpf)
}

// GetPrompt resolves prompt text from positional args, file, or stdin.
func GetPrompt(args []string, filePath string, stdin io.Reader) (string, error) {
	// Priority 1: Positional argument
	if len(args) > 0 {
		text := strings.TrimSpace(args[0])
		if text != "" {
			return text, nil
		}
	}

	// Priority 2: File
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

	// Priority 3: Stdin (only if not a terminal)
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

// HandleSDKError maps Tencent Cloud SDK errors to CLI error responses.
func HandleSDKError(cmd *cobra.Command, err error) error {
	msg := err.Error()

	if strings.Contains(msg, "AuthFailure") {
		return common.WriteError(cmd, "invalid_api_key", msg)
	}
	if strings.Contains(msg, "RequestLimitExceeded") {
		return common.WriteError(cmd, "rate_limit", msg)
	}
	if strings.Contains(msg, "InvalidParameter") || strings.Contains(msg, "InvalidParameterValue") {
		return common.WriteError(cmd, "invalid_request", msg)
	}
	if strings.Contains(msg, "OperationDenied") {
		return common.WriteError(cmd, "content_policy", msg)
	}
	if strings.Contains(msg, "FailedOperation.JobNotExist") || strings.Contains(msg, "FailedOperation.JobNotFound") {
		return common.WriteError(cmd, "task_not_found", msg)
	}
	if strings.Contains(msg, "FailedOperation") {
		return common.WriteError(cmd, "api_error", msg)
	}
	if strings.Contains(msg, "InternalError") {
		return common.WriteError(cmd, "server_error", msg)
	}
	if strings.Contains(msg, "timeout") {
		return common.WriteError(cmd, "timeout", msg)
	}
	if strings.Contains(msg, "connection") || strings.Contains(msg, "no such host") {
		return common.WriteError(cmd, "connection_error", msg)
	}

	return common.WriteError(cmd, "api_error", msg)
}

// IsURL checks if the given string is a URL.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ResolveImageBase64 reads a local file and returns its base64 encoding.
func ResolveImageBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DownloadFile downloads a URL to the given output path and returns the absolute path.
func DownloadFile(cmd *cobra.Command, url, output string) error {
	// Create output directory if needed
	dir := filepath.Dir(output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create directory: %s", err.Error()))
		}
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download file: %s", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %d", resp.StatusCode))
	}

	outFile, err := os.Create(output)
	if err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create file: %s", err.Error()))
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot write file: %s", err.Error()))
	}

	return nil
}

// AbsPath returns the absolute path, falling back to the input if it fails.
func AbsPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
