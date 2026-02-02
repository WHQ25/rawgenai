package audio

import (
	"encoding/json"
	"strings"
	"testing"
)

// SFX Tests
func TestSfx_MissingPrompt(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sfx")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestSfx_InvalidDuration(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sfx", "test", "-d", "0.1")

	if err == nil {
		t.Fatal("expected error for invalid duration")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_duration" {
		t.Errorf("expected error code 'invalid_duration', got: %s", errorObj["code"])
	}
}

func TestSfx_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sfx", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

// TTS Tests
func TestTTS_MissingPrompt(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "tts", "-v", "Maya")

	if err == nil {
		t.Fatal("expected error for missing prompt")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_prompt" {
		t.Errorf("expected error code 'missing_prompt', got: %s", errorObj["code"])
	}
}

func TestTTS_MissingVoice(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "tts", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing voice")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_voice" {
		t.Errorf("expected error code 'missing_voice', got: %s", errorObj["code"])
	}
}

func TestTTS_InvalidVoice(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "tts", "test", "-v", "InvalidVoice")

	if err == nil {
		t.Fatal("expected error for invalid voice")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_voice" {
		t.Errorf("expected error code 'invalid_voice', got: %s", errorObj["code"])
	}
}

// STS Tests
func TestSTS_MissingInput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sts", "-v", "Maya")

	if err == nil {
		t.Fatal("expected error for missing input")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_input" {
		t.Errorf("expected error code 'missing_input', got: %s", errorObj["code"])
	}
}

func TestSTS_MissingVoice(t *testing.T) {
	audioFile := createTempFile(t, "test.mp3", "fake audio")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sts", "-i", audioFile)

	if err == nil {
		t.Fatal("expected error for missing voice")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_voice" {
		t.Errorf("expected error code 'missing_voice', got: %s", errorObj["code"])
	}
}

func TestSTS_InputNotFound(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "sts", "-i", "/nonexistent/audio.mp3", "-v", "Maya")

	if err == nil {
		t.Fatal("expected error for input not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "input_not_found" {
		t.Errorf("expected error code 'input_not_found', got: %s", errorObj["code"])
	}
}

// Dubbing Tests
func TestDubbing_MissingInput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "dubbing", "-l", "es")

	if err == nil {
		t.Fatal("expected error for missing input")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_input" {
		t.Errorf("expected error code 'missing_input', got: %s", errorObj["code"])
	}
}

func TestDubbing_MissingLang(t *testing.T) {
	audioFile := createTempFile(t, "test.mp3", "fake audio")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "dubbing", "-i", audioFile)

	if err == nil {
		t.Fatal("expected error for missing lang")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_lang" {
		t.Errorf("expected error code 'missing_lang', got: %s", errorObj["code"])
	}
}

func TestDubbing_InvalidLang(t *testing.T) {
	audioFile := createTempFile(t, "test.mp3", "fake audio")

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "dubbing", "-i", audioFile, "-l", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid lang")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_lang" {
		t.Errorf("expected error code 'invalid_lang', got: %s", errorObj["code"])
	}
}

// Isolation Tests
func TestIsolation_MissingInput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "isolation")

	if err == nil {
		t.Fatal("expected error for missing input")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_input" {
		t.Errorf("expected error code 'missing_input', got: %s", errorObj["code"])
	}
}

func TestIsolation_InputNotFound(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "isolation", "-i", "/nonexistent/audio.mp3")

	if err == nil {
		t.Fatal("expected error for input not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "input_not_found" {
		t.Errorf("expected error code 'input_not_found', got: %s", errorObj["code"])
	}
}

// Status Tests
func TestStatus_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "status")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

// Download Tests
func TestDownload_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "-o", "output.mp3")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestDownload_InvalidOutputExtension(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id", "-o", "output.txt")

	if err == nil {
		t.Fatal("expected error for invalid output extension")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_output" {
		t.Errorf("expected error code 'invalid_output', got: %s", errorObj["code"])
	}
}

// Delete Tests
func TestDelete_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "delete")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
	}
}
