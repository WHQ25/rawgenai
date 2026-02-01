package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrorInfo contains error details for JSON response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Success bool       `json:"success"`
	Error   *ErrorInfo `json:"error"`
}

// WriteError writes a JSON error response to stderr and returns an error
func WriteError(cmd *cobra.Command, code, message string) error {
	resp := ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	output, _ := json.Marshal(resp)
	fmt.Fprintln(cmd.ErrOrStderr(), string(output))
	return errors.New(code)
}

// WriteSuccess writes a JSON success response to stdout
func WriteSuccess(cmd *cobra.Command, data any) error {
	output, _ := json.Marshal(data)
	fmt.Fprintln(cmd.OutOrStdout(), string(output))
	return nil
}
