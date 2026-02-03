package image

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/kling/video"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	output    string
	index     int
	watermark bool
}

func newDownloadCmd() *cobra.Command {
	flags := &downloadFlags{}

	cmd := &cobra.Command{
		Use:           "download <task_id>",
		Short:         "Download a generated image",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().IntVar(&flags.index, "index", 0, "Image index (0-based)")
	cmd.Flags().BoolVar(&flags.watermark, "watermark", false, "Download watermarked image")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string, flags *downloadFlags) error {
	// Validate task ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_task_id", "task ID is required")
	}
	taskID := args[0]

	// Validate output
	if strings.TrimSpace(flags.output) == "" {
		return common.WriteError(cmd, "missing_output", "output file path is required (-o)")
	}

	// Validate index
	if flags.index < 0 {
		return common.WriteError(cmd, "invalid_index", "index must be >= 0")
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := video.GenerateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	downloadURL, err := getDownloadURL(token, taskID, flags.index, flags.watermark)
	if err != nil {
		return common.WriteError(cmd, "download_error", err.Error())
	}

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("cannot download file: %s", err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "download_error", fmt.Sprintf("download failed with status: %d", resp.StatusCode))
	}

	// Create output directory if needed
	dir := filepath.Dir(flags.output)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create directory: %s", err.Error()))
		}
	}

	// Create output file
	outFile, err := os.Create(flags.output)
	if err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot create file: %s", err.Error()))
	}
	defer outFile.Close()

	// Copy data
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return common.WriteError(cmd, "write_error", fmt.Sprintf("cannot write file: %s", err.Error()))
	}

	absPath, err := filepath.Abs(flags.output)
	if err != nil {
		absPath = flags.output
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success": true,
		"task_id": taskID,
		"file":    absPath,
	})
}

func getDownloadURL(token, taskID string, index int, watermark bool) (string, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", video.GetKlingAPIBase()+"/v1/images/generations/"+taskID, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %s", err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot get status: %s", err.Error())
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response: %s", err.Error())
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskStatus    string `json:"task_status"`
			TaskStatusMsg string `json:"task_status_msg"`
			TaskResult    *struct {
				Images []struct {
					URL          string `json:"url"`
					WatermarkURL string `json:"watermark_url"`
				} `json:"images"`
			} `json:"task_result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("cannot parse response: %s", err.Error())
	}

	if result.Code != 0 {
		return "", fmt.Errorf("kling error: %s", result.Message)
	}

	if result.Data == nil {
		return "", fmt.Errorf("no data in response")
	}

	if result.Data.TaskStatus == "failed" {
		msg := result.Data.TaskStatusMsg
		if msg == "" {
			msg = "image generation failed"
		}
		return "", fmt.Errorf(msg)
	}

	if result.Data.TaskStatus != "succeed" {
		return "", fmt.Errorf("task is not finished (status: %s)", result.Data.TaskStatus)
	}

	if result.Data.TaskResult == nil || len(result.Data.TaskResult.Images) == 0 {
		return "", fmt.Errorf("no images in result")
	}

	if index >= len(result.Data.TaskResult.Images) {
		return "", fmt.Errorf("index out of range")
	}

	img := result.Data.TaskResult.Images[index]
	if watermark {
		if img.WatermarkURL == "" {
			return "", fmt.Errorf("watermark URL not available")
		}
		return img.WatermarkURL, nil
	}

	if img.URL == "" {
		return "", fmt.Errorf("image URL not available")
	}

	return img.URL, nil
}
