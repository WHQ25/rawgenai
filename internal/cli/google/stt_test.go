package google

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

func TestSTT_MissingInput(t *testing.T) {
	cmd := newSTTCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "missing_input" {
		t.Errorf("expected error code 'missing_input', got '%s'", resp.Error.Code)
	}
}

func TestSTT_FileNotFound(t *testing.T) {
	cmd := newSTTCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"nonexistent.mp3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "file_not_found" {
		t.Errorf("expected error code 'file_not_found', got '%s'", resp.Error.Code)
	}
}

func TestSTT_UnsupportedFormat(t *testing.T) {
	// Create a temp file with unsupported extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xyz")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := newSTTCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{tmpFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "unsupported_format" {
		t.Errorf("expected error code 'unsupported_format', got '%s'", resp.Error.Code)
	}
}

func TestSTT_MissingAPIKey(t *testing.T) {
	// Create a temp file with supported extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(tmpFile, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Ensure API key is not set
	oldGemini := os.Getenv("GEMINI_API_KEY")
	oldGoogle := os.Getenv("GOOGLE_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GOOGLE_API_KEY")
	defer func() {
		if oldGemini != "" {
			os.Setenv("GEMINI_API_KEY", oldGemini)
		}
		if oldGoogle != "" {
			os.Setenv("GOOGLE_API_KEY", oldGoogle)
		}
	}()

	cmd := newSTTCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{tmpFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error.Code != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got '%s'", resp.Error.Code)
	}
}

func TestSTT_FileFlag(t *testing.T) {
	// Create a temp file with supported extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(tmpFile, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Ensure API key is not set
	oldGemini := os.Getenv("GEMINI_API_KEY")
	oldGoogle := os.Getenv("GOOGLE_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GOOGLE_API_KEY")
	defer func() {
		if oldGemini != "" {
			os.Setenv("GEMINI_API_KEY", oldGemini)
		}
		if oldGoogle != "" {
			os.Setenv("GOOGLE_API_KEY", oldGoogle)
		}
	}()

	cmd := newSTTCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--file", tmpFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error (missing API key), got nil")
	}

	var resp common.ErrorResponse
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	// Should reach API key validation (file was read successfully)
	if resp.Error.Code != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got '%s'", resp.Error.Code)
	}
}

func TestSTT_FlagDefaults(t *testing.T) {
	cmd := newSTTCmd()

	// Check default values
	timestamps, _ := cmd.Flags().GetBool("timestamps")
	if timestamps != false {
		t.Errorf("expected default timestamps false, got %v", timestamps)
	}

	speakers, _ := cmd.Flags().GetBool("speakers")
	if speakers != false {
		t.Errorf("expected default speakers false, got %v", speakers)
	}

	model, _ := cmd.Flags().GetString("model")
	if model != "flash" {
		t.Errorf("expected default model 'flash', got '%s'", model)
	}
}

func TestSTT_SupportedFormats(t *testing.T) {
	formats := []string{".mp3", ".wav", ".flac", ".ogg", ".aac", ".m4a", ".webm"}
	for _, format := range formats {
		if _, ok := sttSupportedFormats[format]; !ok {
			t.Errorf("expected format %s to be supported", format)
		}
	}
}

func TestBuildTranscriptionPrompt(t *testing.T) {
	tests := []struct {
		name       string
		language   string
		timestamps bool
		speakers   bool
		contains   []string
	}{
		{
			name:       "basic",
			language:   "",
			timestamps: false,
			speakers:   false,
			contains:   []string{"Transcribe", "JSON", "text"},
		},
		{
			name:       "with language",
			language:   "en",
			timestamps: false,
			speakers:   false,
			contains:   []string{"en language"},
		},
		{
			name:       "with timestamps",
			language:   "",
			timestamps: true,
			speakers:   false,
			contains:   []string{"start", "end", "timestamps"},
		},
		{
			name:       "with speakers",
			language:   "",
			timestamps: false,
			speakers:   true,
			contains:   []string{"speaker", "Speaker 1"},
		},
		{
			name:       "full",
			language:   "zh",
			timestamps: true,
			speakers:   true,
			contains:   []string{"zh language", "start", "end", "speaker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildTranscriptionPrompt(tt.language, tt.timestamps, tt.speakers)
			for _, s := range tt.contains {
				if !bytes.Contains([]byte(prompt), []byte(s)) {
					t.Errorf("expected prompt to contain '%s', got: %s", s, prompt)
				}
			}
		})
	}
}
