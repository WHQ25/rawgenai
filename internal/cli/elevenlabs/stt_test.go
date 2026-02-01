package elevenlabs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSTT_MissingInput(t *testing.T) {
	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd)

	if err == nil {
		t.Fatal("expected error for missing input")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_input" {
		t.Errorf("expected error code 'missing_input', got: %s", errorObj["code"])
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

func TestSTT_InvalidAudioFormat(t *testing.T) {
	// Create temp file with invalid extension
	tmpFile, err := os.CreateTemp("", "stt_test_*.xyz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name())

	if err == nil {
		t.Fatal("expected error for invalid audio format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_audio" {
		t.Errorf("expected error code 'invalid_audio', got: %s", errorObj["code"])
	}
}

func TestSTT_InvalidModel(t *testing.T) {
	// Create temp audio file
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name(), "-m", "invalid_model")

	if err == nil {
		t.Fatal("expected error for invalid model")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_model" {
		t.Errorf("expected error code 'invalid_model', got: %s", errorObj["code"])
	}
}

func TestSTT_InvalidTimestamps(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name(), "--timestamps", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid timestamps")
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

func TestSTT_SpeakersWithoutDiarize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name(), "--speakers", "3")

	if err == nil {
		t.Fatal("expected error for speakers without diarize")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_parameter" {
		t.Errorf("expected error code 'invalid_parameter', got: %s", errorObj["code"])
	}
	if !strings.Contains(errorObj["message"].(string), "--diarize") {
		t.Errorf("expected error message to mention --diarize, got: %s", errorObj["message"])
	}
}

func TestSTT_SpeakersOutOfRange(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name(), "--diarize", "--speakers", "50")

	if err == nil {
		t.Fatal("expected error for speakers out of range")
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

func TestSTT_MissingAPIKey(t *testing.T) {
	t.Setenv("ELEVENLABS_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, tmpFile.Name())

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

func TestSTT_ValidFlags(t *testing.T) {
	cmd := newSTTCmd()

	if cmd.Flag("file") == nil {
		t.Error("expected --file flag")
	}
	if cmd.Flag("model") == nil {
		t.Error("expected --model flag")
	}
	if cmd.Flag("language") == nil {
		t.Error("expected --language flag")
	}
	if cmd.Flag("diarize") == nil {
		t.Error("expected --diarize flag")
	}
	if cmd.Flag("speakers") == nil {
		t.Error("expected --speakers flag")
	}
	if cmd.Flag("timestamps") == nil {
		t.Error("expected --timestamps flag")
	}
	if cmd.Flag("output") == nil {
		t.Error("expected --output flag")
	}
}

func TestSTT_DefaultValues(t *testing.T) {
	cmd := newSTTCmd()

	if cmd.Flag("model").DefValue != "scribe_v1" {
		t.Errorf("expected default model 'scribe_v1', got: %s", cmd.Flag("model").DefValue)
	}
	if cmd.Flag("timestamps").DefValue != "word" {
		t.Errorf("expected default timestamps 'word', got: %s", cmd.Flag("timestamps").DefValue)
	}
	if cmd.Flag("diarize").DefValue != "false" {
		t.Errorf("expected default diarize 'false', got: %s", cmd.Flag("diarize").DefValue)
	}
}

func TestSTT_FromFileFlag(t *testing.T) {
	t.Setenv("ELEVENLABS_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "stt_test_*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake audio content")
	tmpFile.Close()

	cmd := newSTTCmd()
	_, stderr, err := executeCommand(cmd, "-f", tmpFile.Name())

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

func TestSTT_FromStdin(t *testing.T) {
	t.Setenv("ELEVENLABS_API_KEY", "")

	cmd := newSTTCmd()
	cmd.SetIn(strings.NewReader("fake audio from stdin"))

	_, stderr, err := executeCommand(cmd)

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

func TestSTT_FormatSRTTime(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0, "00:00:00,000"},
		{1.5, "00:00:01,500"},
		{61.123, "00:01:01,123"},
		{3661.999, "01:01:01,999"},
	}

	for _, tt := range tests {
		result := formatSRTTime(tt.seconds)
		if result != tt.expected {
			t.Errorf("formatSRTTime(%f) = %q, want %q", tt.seconds, result, tt.expected)
		}
	}
}

func TestGenerateSTTSRT(t *testing.T) {
	tests := []struct {
		name     string
		response sttAPIResponse
		expected string
	}{
		{
			name: "filters spacing tokens",
			response: sttAPIResponse{
				Text: "Hello world",
				Words: []sttAPIWord{
					{Text: "Hello", Start: 0.1, End: 0.5, Type: "word"},
					{Text: " ", Start: 0.5, End: 0.6, Type: "spacing"},
					{Text: "world", Start: 0.6, End: 1.0, Type: "word"},
				},
			},
			expected: "1\n00:00:00,100 --> 00:00:01,000\nHello world\n\n",
		},
		{
			name: "empty words returns empty string",
			response: sttAPIResponse{
				Text:  "Hello",
				Words: []sttAPIWord{},
			},
			expected: "",
		},
		{
			name: "only spacing tokens returns empty",
			response: sttAPIResponse{
				Text: " ",
				Words: []sttAPIWord{
					{Text: " ", Start: 0.0, End: 0.1, Type: "spacing"},
				},
			},
			expected: "",
		},
		{
			name: "segments long text",
			response: sttAPIResponse{
				Text: "One two three four five six seven eight nine ten eleven",
				Words: []sttAPIWord{
					{Text: "One", Start: 0.0, End: 0.1, Type: "word"},
					{Text: "two", Start: 0.2, End: 0.3, Type: "word"},
					{Text: "three", Start: 0.4, End: 0.5, Type: "word"},
					{Text: "four", Start: 0.6, End: 0.7, Type: "word"},
					{Text: "five", Start: 0.8, End: 0.9, Type: "word"},
					{Text: "six", Start: 1.0, End: 1.1, Type: "word"},
					{Text: "seven", Start: 1.2, End: 1.3, Type: "word"},
					{Text: "eight", Start: 1.4, End: 1.5, Type: "word"},
					{Text: "nine", Start: 1.6, End: 1.7, Type: "word"},
					{Text: "ten", Start: 1.8, End: 1.9, Type: "word"},
					{Text: "eleven", Start: 2.0, End: 2.1, Type: "word"},
				},
			},
			// First segment: 10 words, second segment: 1 word
			expected: "1\n00:00:00,000 --> 00:00:01,900\nOne two three four five six seven eight nine ten\n\n2\n00:00:02,000 --> 00:00:02,100\neleven\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSTTSRT(tt.response)
			if result != tt.expected {
				t.Errorf("generateSTTSRT() =\n%q\nwant\n%q", result, tt.expected)
			}
		})
	}
}
