package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/cli/minimax/shared"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type uploadFlags struct {
	file    string
	purpose string
}

var validUploadPurposes = map[string]bool{
	"voice_clone": true,
}

func newUploadCmd() *cobra.Command {
	flags := &uploadFlags{}

	cmd := &cobra.Command{
		Use:           "upload",
		Short:         "Upload audio file for voice cloning",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "Audio file path (required)")
	cmd.Flags().StringVar(&flags.purpose, "purpose", "voice_clone", "Purpose: voice_clone")

	return cmd
}

func runUpload(cmd *cobra.Command, flags *uploadFlags) error {
	if flags.file == "" {
		return common.WriteError(cmd, "missing_file", "audio file is required")
	}
	if !validUploadPurposes[flags.purpose] {
		return common.WriteError(cmd, "invalid_purpose", "purpose must be voice_clone")
	}

	if _, err := os.Stat(flags.file); os.IsNotExist(err) {
		return common.WriteError(cmd, "file_not_found", fmt.Sprintf("file not found: %s", flags.file))
	}

	apiKey := shared.GetMinimaxAPIKey()
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("MINIMAX_API_KEY"))
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("purpose", flags.purpose); err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot write purpose: %s", err.Error()))
	}

	file, err := os.Open(flags.file)
	if err != nil {
		return common.WriteError(cmd, "file_read_error", fmt.Sprintf("cannot open file: %s", err.Error()))
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(flags.file))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create form file: %s", err.Error()))
	}

	if _, err := io.Copy(part, file); err != nil {
		return common.WriteError(cmd, "file_read_error", fmt.Sprintf("cannot read file: %s", err.Error()))
	}

	if err := writer.Close(); err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot finalize form: %s", err.Error()))
	}

	req, err := http.NewRequest("POST", shared.MinimaxAPIBase+"/v1/files/upload", body)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := shared.DoRequest(req)
	if err != nil {
		return common.WriteError(cmd, "request_error", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	var apiResp struct {
		File struct {
			FileID   int64  `json:"file_id"`
			Bytes    int64  `json:"bytes"`
			Filename string `json:"filename"`
			Purpose  string `json:"purpose"`
		} `json:"file"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return common.WriteError(cmd, "api_error", fmt.Sprintf("api error %d: %s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMsg))
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":  true,
		"file_id":  apiResp.File.FileID,
		"bytes":    apiResp.File.Bytes,
		"filename": apiResp.File.Filename,
		"purpose":  apiResp.File.Purpose,
	})
}
