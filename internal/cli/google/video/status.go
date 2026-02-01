package video

import (
	"context"
	"os"
	"strings"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

type statusResponse struct {
	Success     bool    `json:"success"`
	OperationID string  `json:"operation_id"`
	Status      string  `json:"status"`
	Progress    float64 `json:"progress,omitempty"`
	Error       string  `json:"error_message,omitempty"`
}

var statusCmd = newStatusCmd()

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status <operation_id>",
		Short:         "Get video generation status",
		Long:          "Query the status of a video generation operation.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	operationID := strings.TrimSpace(args[0])
	if operationID == "" {
		return common.WriteError(cmd, "missing_operation_id", "operation_id is required")
	}

	// Check API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is not set")
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return common.WriteError(cmd, "client_error", err.Error())
	}

	// Create operation reference with just the name
	opRef := &genai.GenerateVideosOperation{
		Name: operationID,
	}

	// Get operation status
	op, err := client.Operations.GetVideosOperation(ctx, opRef, nil)
	if err != nil {
		return handleAPIError(cmd, err)
	}

	result := statusResponse{
		Success:     true,
		OperationID: op.Name,
	}

	if op.Done {
		// Check if there was an error
		if op.Error != nil {
			if msg, ok := op.Error["message"].(string); ok && msg != "" {
				result.Status = "failed"
				result.Error = msg
			} else {
				result.Status = "completed"
				result.Progress = 1.0
			}
		} else {
			result.Status = "completed"
			result.Progress = 1.0
		}
	} else {
		result.Status = "running"
		// Calculate progress if metadata is available
		if op.Metadata != nil {
			if progress, ok := op.Metadata["progress"].(float64); ok {
				result.Progress = progress
			}
		}
	}

	return common.WriteSuccess(cmd, result)
}
