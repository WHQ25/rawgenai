package elevenlabs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeVoiceListCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestVoiceList_InvalidPageSize(t *testing.T) {
	tests := []struct {
		name     string
		pageSize string
	}{
		{"too_low", "0"},
		{"too_high", "150"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVoiceListCmd()
			_, stderr, err := executeVoiceListCommand(cmd, "--page-size", tt.pageSize)

			if err == nil {
				t.Fatal("expected error for invalid page size")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_page_size" {
				t.Errorf("expected error code 'invalid_page_size', got '%s'", errorObj["code"])
			}
		})
	}
}

func TestVoiceList_InvalidVoiceType(t *testing.T) {
	cmd := newVoiceListCmd()
	_, stderr, err := executeVoiceListCommand(cmd, "--voice-type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid voice type")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_voice_type" {
		t.Errorf("expected error code 'invalid_voice_type', got '%s'", errorObj["code"])
	}
}

func TestVoiceList_InvalidCategory(t *testing.T) {
	cmd := newVoiceListCmd()
	_, stderr, err := executeVoiceListCommand(cmd, "--category", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid category")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_category" {
		t.Errorf("expected error code 'invalid_category', got '%s'", errorObj["code"])
	}
}

func TestVoiceList_InvalidSort(t *testing.T) {
	cmd := newVoiceListCmd()
	_, stderr, err := executeVoiceListCommand(cmd, "--sort", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid sort")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_sort" {
		t.Errorf("expected error code 'invalid_sort', got '%s'", errorObj["code"])
	}
}

func TestVoiceList_InvalidSortDir(t *testing.T) {
	cmd := newVoiceListCmd()
	_, stderr, err := executeVoiceListCommand(cmd, "--sort-dir", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid sort direction")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_sort_dir" {
		t.Errorf("expected error code 'invalid_sort_dir', got '%s'", errorObj["code"])
	}
}

func TestVoiceList_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newVoiceListCmd()
	_, stderr, err := executeVoiceListCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got '%s'", errorObj["code"])
	}
}

func TestVoiceList_ValidFlags(t *testing.T) {
	cmd := newVoiceListCmd()

	flags := []string{"search", "voice-type", "category", "page-size", "page-token", "sort", "sort-dir", "collection-id", "voice-ids", "total-count"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag '%s' not found", flag)
		}
	}
}

func TestVoiceList_DefaultValues(t *testing.T) {
	cmd := newVoiceListCmd()

	tests := []struct {
		flag     string
		expected string
	}{
		{"page-size", "10"},
		{"total-count", "true"},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.flag)
		if flag == nil {
			t.Errorf("flag '%s' not found", tt.flag)
			continue
		}
		if flag.DefValue != tt.expected {
			t.Errorf("flag '%s' default: expected '%s', got '%s'", tt.flag, tt.expected, flag.DefValue)
		}
	}
}

func TestVoiceList_ValidVoiceTypes(t *testing.T) {
	// This test just verifies valid voice types don't cause validation errors
	// It doesn't make API calls
	cmd := newVoiceListCmd()

	validTypes := []string{"personal", "community", "default", "workspace", "non-default", "saved"}
	for _, voiceType := range validTypes {
		// Just check the flag can be set without error
		if err := cmd.Flags().Set("voice-type", voiceType); err != nil {
			t.Errorf("failed to set voice-type to '%s': %v", voiceType, err)
		}
	}
}

func TestVoiceList_ValidCategories(t *testing.T) {
	// This test just verifies valid categories don't cause validation errors
	cmd := newVoiceListCmd()

	validCategories := []string{"premade", "cloned", "generated", "professional"}
	for _, category := range validCategories {
		if err := cmd.Flags().Set("category", category); err != nil {
			t.Errorf("failed to set category to '%s': %v", category, err)
		}
	}
}

func TestVoiceList_ValidSortOptions(t *testing.T) {
	// This test just verifies valid sort options don't cause validation errors
	cmd := newVoiceListCmd()

	validSorts := []string{"created_at_unix", "name"}
	for _, sort := range validSorts {
		if err := cmd.Flags().Set("sort", sort); err != nil {
			t.Errorf("failed to set sort to '%s': %v", sort, err)
		}
	}
}
