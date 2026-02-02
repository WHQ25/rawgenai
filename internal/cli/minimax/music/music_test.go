package music

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMusic_MissingLyrics(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "-o", "out.mp3")

	if err == nil {
		t.Fatal("expected error for missing lyrics")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_lyrics" {
		t.Errorf("expected error code missing_lyrics, got: %v", errObj["code"])
	}
}

func TestMusic_MissingOutput(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "[verse]\nHello world")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "missing_output" {
		t.Errorf("expected error code missing_output, got: %v", errObj["code"])
	}
}

func TestMusic_InvalidFormat(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "[verse]\nHello", "-o", "out.mp3", "-f", "ogg")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_format" {
		t.Errorf("expected error code invalid_format, got: %v", errObj["code"])
	}
}

func TestMusic_InvalidSampleRate(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "[verse]\nHello", "-o", "out.mp3", "--sample-rate", "48000")

	if err == nil {
		t.Fatal("expected error for invalid sample rate")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_sample_rate" {
		t.Errorf("expected error code invalid_sample_rate, got: %v", errObj["code"])
	}
}

func TestMusic_InvalidBitrate(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "[verse]\nHello", "-o", "out.mp3", "--bitrate", "192000")

	if err == nil {
		t.Fatal("expected error for invalid bitrate")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "invalid_bitrate" {
		t.Errorf("expected error code invalid_bitrate, got: %v", errObj["code"])
	}
}

func TestMusic_FormatMismatch(t *testing.T) {
	cmd := newCreateCmd()
	_, stderr, err := executeCommand(cmd, "[verse]\nHello", "-o", "out.wav", "-f", "mp3")

	if err == nil {
		t.Fatal("expected error for format mismatch")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errObj := resp["error"].(map[string]any)
	if errObj["code"] != "format_mismatch" {
		t.Errorf("expected error code format_mismatch, got: %v", errObj["code"])
	}
}

func TestMusic_AllFlags(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	expectedFlags := []string{
		"output", "lyrics-file", "prompt", "stream", "play",
		"format", "sample-rate", "bitrate",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}
}

func TestMusic_ShortFlags(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	shortFlags := map[string]string{
		"o": "output",
		"p": "prompt",
		"f": "format",
	}

	for short, long := range shortFlags {
		flag := flags.Lookup(long)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", long)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("expected flag --%s to have shorthand -%s, got -%s", long, short, flag.Shorthand)
		}
	}
}

func TestMusic_DefaultValues(t *testing.T) {
	cmd := newCreateCmd()
	flags := cmd.Flags()

	defaults := map[string]string{
		"format":      "mp3",
		"sample-rate": "44100",
		"bitrate":     "256000",
	}

	for name, expected := range defaults {
		flag := flags.Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected flag --%s default to be %s, got %s", name, expected, flag.DefValue)
		}
	}
}
