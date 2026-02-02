package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVideo2Video_MissingPrompt(t *testing.T) {
	videoFile := createTempFile(t, "test.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newVideo2VideoCmd())
	_, stderr, err := executeCommand(cmd, "video2video", "-v", videoFile)

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

func TestVideo2Video_MissingVideo(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newVideo2VideoCmd())
	_, stderr, err := executeCommand(cmd, "video2video", "test prompt")

	if err == nil {
		t.Fatal("expected error for missing video")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_video" {
		t.Errorf("expected error code 'missing_video', got: %s", errorObj["code"])
	}
}

func TestVideo2Video_VideoNotFound(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newVideo2VideoCmd())
	_, stderr, err := executeCommand(cmd, "video2video", "test", "-v", "/nonexistent/video.mp4")

	if err == nil {
		t.Fatal("expected error for video not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "video_not_found" {
		t.Errorf("expected error code 'video_not_found', got: %s", errorObj["code"])
	}
}

func TestVideo2Video_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)
	videoFile := createTempFile(t, "test.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newVideo2VideoCmd())
	_, stderr, err := executeCommand(cmd, "video2video", "test", "-v", videoFile)

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

func TestVideo2Video_AllFlags(t *testing.T) {
	cmd := newVideo2VideoCmd()

	expectedFlags := []string{"video", "ratio", "seed", "ref-image", "prompt-file", "public-figure"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}
