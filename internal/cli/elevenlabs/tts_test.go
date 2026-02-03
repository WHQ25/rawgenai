package elevenlabs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "-o", "output.mp3")

	if err == nil {
		t.Fatal("expected error for missing text")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_text" {
		t.Errorf("expected error code 'missing_text', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello world")

	if err == nil {
		t.Fatal("expected error for missing output")
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

func TestTTS_InvalidSpeed(t *testing.T) {
	tests := []struct {
		name  string
		speed string
	}{
		{"too slow", "0.1"},
		{"too fast", "5.0"},
		{"negative", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--speed", tt.speed)

			if err == nil {
				t.Fatal("expected error for invalid speed")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_speed" {
				t.Errorf("expected error code 'invalid_speed', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_InvalidStability(t *testing.T) {
	tests := []struct {
		name      string
		stability string
	}{
		{"negative", "-0.1"},
		{"too high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--stability", tt.stability)

			if err == nil {
				t.Fatal("expected error for invalid stability")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_stability" {
				t.Errorf("expected error code 'invalid_stability', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_InvalidSimilarity(t *testing.T) {
	tests := []struct {
		name       string
		similarity string
	}{
		{"negative", "-0.1"},
		{"too high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--similarity", tt.similarity)

			if err == nil {
				t.Fatal("expected error for invalid similarity")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_similarity" {
				t.Errorf("expected error code 'invalid_similarity', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_InvalidStyle(t *testing.T) {
	tests := []struct {
		name  string
		style string
	}{
		{"negative", "-0.1"},
		{"too high", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--style", tt.style)

			if err == nil {
				t.Fatal("expected error for invalid style")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_style" {
				t.Errorf("expected error code 'invalid_style', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_InvalidFormat(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "-f", "invalid_format")

	if err == nil {
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

func TestTTS_InvalidTextNormalization(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--text-normalization", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid text normalization")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_text_normalization" {
		t.Errorf("expected error code 'invalid_text_normalization', got: %s", errorObj["code"])
	}
}

func TestTTS_ValidTextNormalization(t *testing.T) {
	validValues := []string{"auto", "on", "off"}

	for _, value := range validValues {
		t.Run(value, func(t *testing.T) {
			common.SetupNoConfigEnv(t)
			t.Setenv("ELEVENLABS_API_KEY", "")

			cmd := newTTSCmd()
			_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--text-normalization", value)

			if err == nil {
				t.Fatal("expected error (missing api key), got success")
			}

			var resp map[string]any
			if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
				t.Fatalf("expected JSON error output, got: %s", stderr)
			}

			// Should pass validation and reach API key check
			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "missing_api_key" {
				t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
			}
		})
	}
}

func TestTTS_StreamFlag(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--stream")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should pass validation and reach API key check
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3")

	if err == nil {
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

func TestTTS_ValidFlags(t *testing.T) {
	cmd := newTTSCmd()

	expectedFlags := []string{
		"output", "file", "voice", "model", "format", "language",
		"stability", "similarity", "style", "speed",
		"speaker-boost", "text-normalization", "stream", "speak",
	}

	for _, flag := range expectedFlags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestTTS_SpeakWithoutOutput(t *testing.T) {
	// --speak without -o should not trigger missing_output error
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "--speak")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	// Should reach API key check, not missing_output
	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestTTS_DefaultValues(t *testing.T) {
	cmd := newTTSCmd()

	defaults := map[string]string{
		"voice":              "Rachel",
		"model":              "eleven_multilingual_v2",
		"format":             "mp3_44100_128",
		"language":           "",
		"stability":          "0.5",
		"similarity":         "0.75",
		"style":              "0",
		"speed":              "1",
		"speaker-boost":      "true",
		"text-normalization": "auto",
		"stream":             "false",
	}

	for flag, expected := range defaults {
		if cmd.Flag(flag).DefValue != expected {
			t.Errorf("expected default %s '%s', got: %s", flag, expected, cmd.Flag(flag).DefValue)
		}
	}
}

func TestTTS_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tts_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("Hello from file")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--file", tmpFile.Name(), "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (file read success), got: %s", errorObj["code"])
	}
}

func TestTTS_FromFileNotFound(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "--file", "/nonexistent/file.txt", "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for file not found")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_text" {
		t.Errorf("expected error code 'missing_text', got: %s", errorObj["code"])
	}
}

func TestTTS_FromStdin(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newTTSCmd()
	cmd.SetIn(strings.NewReader("Hello from stdin"))

	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error (missing api key), got success")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key' (stdin read success), got: %s", errorObj["code"])
	}
}

func TestTTS_ResolveVoiceID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Rachel", "21m00Tcm4TlvDq8ikWAM"},
		{"rachel", "21m00Tcm4TlvDq8ikWAM"},
		{"RACHEL", "21m00Tcm4TlvDq8ikWAM"},
		{"josh", "TxGEqnHWrfWFTfGW9XjX"},
		{"21m00Tcm4TlvDq8ikWAM", "21m00Tcm4TlvDq8ikWAM"}, // Already an ID
		{"custom_voice_id", "custom_voice_id"},             // Unknown voice treated as ID
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := resolveVoiceID(tt.input)
			if result != tt.expected {
				t.Errorf("resolveVoiceID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHandleAPIErrorResponse_QuotaExceeded(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedCode string
	}{
		// 400/401 errors
		{
			name:         "quota exceeded with 401 status",
			statusCode:   401,
			responseBody: `{"detail":{"status":"quota_exceeded","message":"You have 0 credits remaining"}}`,
			expectedCode: "quota_exceeded",
		},
		{
			name:         "quota in message with 401",
			statusCode:   401,
			responseBody: `{"detail":{"status":"error","message":"API quota exceeded"}}`,
			expectedCode: "quota_exceeded",
		},
		{
			name:         "actual invalid api key",
			statusCode:   401,
			responseBody: `{"detail":{"status":"invalid_api_key","message":"Invalid API key"}}`,
			expectedCode: "invalid_api_key",
		},
		{
			name:         "max character limit exceeded",
			statusCode:   400,
			responseBody: `{"detail":{"status":"max_character_limit_exceeded","message":"Text exceeds maximum character limit"}}`,
			expectedCode: "max_character_limit_exceeded",
		},
		{
			name:         "voice not found",
			statusCode:   400,
			responseBody: `{"detail":{"status":"voice_not_found","message":"Voice with ID xxx not found"}}`,
			expectedCode: "voice_not_found",
		},
		// 403 errors
		{
			name:         "subscription required for pro voices",
			statusCode:   403,
			responseBody: `{"detail":{"status":"only_for_creator+","message":"Professional voices require Creator plan or above"}}`,
			expectedCode: "subscription_required",
		},
		// 429 errors
		{
			name:         "rate limit 429",
			statusCode:   429,
			responseBody: `{"detail":{"status":"rate_limit","message":"Too many requests"}}`,
			expectedCode: "rate_limit",
		},
		{
			name:         "too many concurrent requests",
			statusCode:   429,
			responseBody: `{"detail":{"status":"too_many_concurrent_requests","message":"You have exceeded the concurrency limit"}}`,
			expectedCode: "too_many_concurrent_requests",
		},
		{
			name:         "system busy",
			statusCode:   429,
			responseBody: `{"detail":{"status":"system_busy","message":"Our services are experiencing high levels of traffic"}}`,
			expectedCode: "system_busy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTTSCmd()
			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)
			cmd.SetOut(outBuf)
			cmd.SetErr(errBuf)

			resp := &mockHTTPResponse{
				statusCode: tt.statusCode,
				body:       tt.responseBody,
			}

			_ = handleAPIErrorResponse(cmd, resp.toHTTPResponse())

			var result map[string]any
			if err := json.Unmarshal(errBuf.Bytes(), &result); err != nil {
				t.Fatalf("failed to parse error response: %v", err)
			}

			errorObj := result["error"].(map[string]any)
			if errorObj["code"] != tt.expectedCode {
				t.Errorf("expected error code %q, got %q", tt.expectedCode, errorObj["code"])
			}
		})
	}
}

type mockHTTPResponse struct {
	statusCode int
	body       string
}

func (m *mockHTTPResponse) toHTTPResponse() *http.Response {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}
}
