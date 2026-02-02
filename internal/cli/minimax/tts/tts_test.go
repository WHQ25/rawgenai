package tts

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTTS_MissingText(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")
	if err == nil {
		t.Fatal("expected error for missing text")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestTTS_MissingOutput(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello")
	if err == nil {
		t.Fatal("expected error for missing output")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestTTS_InvalidFormat(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--format", "bad")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestTTS_InvalidSpeed(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "Hello", "-o", "out.mp3", "--speed", "3")
	if err == nil {
		t.Fatal("expected error for invalid speed")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestTTSCreate_MissingTextAndFileID(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "create")
	if err == nil {
		t.Fatal("expected error for missing text and file-id")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}

func TestTTSCreate_TextAndFileIDConflict(t *testing.T) {
	cmd := newTTSCmd()
	_, stderr, err := executeCommand(cmd, "create", "Hello", "--file-id", "123")
	if err == nil {
		t.Fatal("expected error for text and file-id conflict")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}
