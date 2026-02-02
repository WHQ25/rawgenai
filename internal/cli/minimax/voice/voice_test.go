package voice

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVoiceList_InvalidType(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "list", "--type", "bad")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestVoiceUpload_MissingFile(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "upload")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestVoiceClone_MissingFileID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "clone", "--voice-id", "voice123")
	if err == nil {
		t.Fatal("expected error for missing file-id")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestVoiceClone_PromptMismatch(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "clone", "--file-id", "123", "--voice-id", "v", "--prompt-audio-id", "999")
	if err == nil {
		t.Fatal("expected error for prompt mismatch")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestVoiceDesign_MissingPrompt(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "design", "--preview-text", "hello", "-o", "out.mp3")
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestVoiceDelete_InvalidType(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "delete", "voice123", "--type", "bad")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}
