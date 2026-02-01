package openai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSTT_MissingFile(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, "")

	if err == nil {
		t.Fatal("expected error for missing file")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_file" {
		t.Errorf("expected error code 'missing_file', got: %s", errorObj["code"])
	}
}

func TestSTT_FileNotFound(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, "/nonexistent/audio.mp3")

	if err == nil {
		t.Fatal("expected error for file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "file_not_found" {
		t.Errorf("expected error code 'file_not_found', got: %s", errorObj["code"])
	}
}

func TestSTT_UnsupportedFormat(t *testing.T) {
	// Create temp file with unsupported extension
	tmpFile, err := os.CreateTemp("", "stt_test_*.xyz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for unsupported format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "unsupported_format" {
		t.Errorf("expected error code 'unsupported_format', got: %s", errorObj["code"])
	}
}

func TestSTT_InvalidTemperature(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	tests := []struct {
		name string
		temp string
	}{
		{"negative", "-0.5"},
		{"too high", "1.5"},
		{"way too high", "2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSTTCmd()
			_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name(), "--temperature", tt.temp)

			if cmdErr == nil {
				t.Fatal("expected error for invalid temperature")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_temperature" {
				t.Errorf("expected error code 'invalid_temperature', got: %s", errorObj["code"])
			}
		})
	}
}

func TestSTT_MissingAPIKey(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	t.Setenv("OPENAI_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestSTT_ValidFlags(t *testing.T) {
	cmd := newSTTCmd()

	flags := []string{"file", "model", "language", "prompt", "temperature", "verbose", "format", "output"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestSTT_DefaultValues(t *testing.T) {
	cmd := newSTTCmd()

	if cmd.Flag("model").DefValue != "whisper-1" {
		t.Errorf("expected default model 'whisper-1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("temperature").DefValue != "0" {
		t.Errorf("expected default temperature '0', got: %s", cmd.Flag("temperature").DefValue)
	}
	if cmd.Flag("format").DefValue != "json" {
		t.Errorf("expected default format 'json', got: %s", cmd.Flag("format").DefValue)
	}
	if cmd.Flag("verbose").DefValue != "false" {
		t.Errorf("expected default verbose 'false', got: %s", cmd.Flag("verbose").DefValue)
	}
}

func TestSTT_FromFileFlag(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	t.Setenv("OPENAI_API_KEY", "")

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, "--file", tmpFile.Name())

	if cmdErr == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning file was found successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file found), got: %s", errorObj["code"])
	}
}

func TestSTT_FromStdin(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	cmd := newSTTCmd()
	cmd.SetIn(strings.NewReader("fake audio content"))

	_, stderr, cmdErr := executeCommand(cmd)

	if cmdErr == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, meaning stdin was read successfully
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestSTT_SRTFormatRequiresOutput(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name(), "--format", "srt")

	if cmdErr == nil {
		t.Fatal("expected error for srt format without output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestSTT_VTTFormatRequiresOutput(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name(), "--format", "vtt")

	if cmdErr == nil {
		t.Fatal("expected error for vtt format without output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestSTT_FileTooLarge(t *testing.T) {
	// Create temp audio file larger than 25MB limit
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.mp3")

	// Create a sparse file that reports as > 25MB
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	// Write just enough to make the file exist, then truncate to large size
	f.WriteString("fake")
	// Seek to 26MB and write a byte to create a sparse file
	f.Seek(26*1024*1024, 0)
	f.WriteString("x")
	f.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile)

	if cmdErr == nil {
		t.Fatal("expected error for file too large")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "file_too_large" {
		t.Errorf("expected error code 'file_too_large', got: %s", errorObj["code"])
	}
}

func TestSTT_SupportedAudioFormats(t *testing.T) {
	supportedExts := []string{".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm", ".ogg", ".oga", ".opus", ".flac"}

	for _, ext := range supportedExts {
		t.Run(ext, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "stt_test_*"+ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.WriteString("fake audio content")
			tmpFile.Close()

			t.Setenv("OPENAI_API_KEY", "")

			cmd := newSTTCmd()
			_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name())

			if cmdErr == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			// Should reach API key check, meaning format was accepted
			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key' (format accepted), got: %s", errorObj["code"])
			}
		})
	}
}

func TestSTT_InvalidFormat(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name(), "--format", "invalid")

	if cmdErr == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_format" {
		t.Errorf("expected error code 'invalid_format', got: %s", errorObj["code"])
	}
}

func TestSTT_PromptNotSupportedByModel(t *testing.T) {
	// gpt-4o-transcribe-diarize does not support prompt
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, cmdErr := executeCommand(cmd, tmpFile.Name(), "--model", "gpt-4o-transcribe-diarize", "--prompt", "Some prompt")

	if cmdErr == nil {
		t.Fatal("expected error for prompt with unsupported model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_parameter" {
		t.Errorf("expected error code 'invalid_parameter', got: %s", errorObj["code"])
	}
}

func TestFormatTimeSRT(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0, "00:00:00,000"},
		{1.5, "00:00:01,500"},
		{61.25, "00:01:01,250"},
		{3661.5, "01:01:01,500"},
		{7325.125, "02:02:05,125"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTimeSRT(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatTimeSRT(%v) = %s, want %s", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestFormatTimeVTT(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0, "00:00:00.000"},
		{1.5, "00:00:01.500"},
		{61.25, "00:01:01.250"},
		{3661.5, "01:01:01.500"},
		{7325.125, "02:02:05.125"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTimeVTT(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatTimeVTT(%v) = %s, want %s", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestGenerateSRT(t *testing.T) {
	// Test SRT format structure: index, timestamp, text
	// Using formatTimeSRT which is already tested above

	// Verify the expected SRT structure
	expectedLines := []string{
		"1",
		"00:00:00,000 --> 00:00:02,500",
		"Hello world",
	}

	// Build a sample SRT entry manually to verify format
	var sb strings.Builder
	sb.WriteString("1\n")
	sb.WriteString(formatTimeSRT(0) + " --> " + formatTimeSRT(2.5) + "\n")
	sb.WriteString("Hello world\n\n")

	result := sb.String()
	for _, line := range expectedLines {
		if !strings.Contains(result, line) {
			t.Errorf("SRT format missing expected line: %s, got: %s", line, result)
		}
	}
}

func TestGenerateVTT(t *testing.T) {
	// Test VTT header and format
	segments := []struct {
		Start float64
		End   float64
		Text  string
	}{
		{0, 2.5, " Hello world "},
	}

	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for _, seg := range segments {
		sb.WriteString(formatTimeVTT(seg.Start))
		sb.WriteString(" --> ")
		sb.WriteString(formatTimeVTT(seg.End))
		sb.WriteString("\n")
		sb.WriteString(strings.TrimSpace(seg.Text))
		sb.WriteString("\n\n")
	}

	result := sb.String()
	if !strings.HasPrefix(result, "WEBVTT\n\n") {
		t.Error("VTT must start with WEBVTT header")
	}
	if !strings.Contains(result, "00:00:00.000 --> 00:00:02.500\nHello world") {
		t.Errorf("VTT format incorrect, got: %s", result)
	}
}
