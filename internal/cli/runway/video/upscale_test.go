package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUpscale_MissingVideo(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newUpscaleCmd())
	_, stderr, err := executeCommand(cmd, "upscale")

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

func TestUpscale_VideoNotFound(t *testing.T) {
	cmd := newTestCmd()
	cmd.AddCommand(newUpscaleCmd())
	_, stderr, err := executeCommand(cmd, "upscale", "-v", "/nonexistent/video.mp4")

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

func TestUpscale_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)
	videoFile := createTempFile(t, "test.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newUpscaleCmd())
	_, stderr, err := executeCommand(cmd, "upscale", "-v", videoFile)

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

func TestUpscale_AllFlags(t *testing.T) {
	cmd := newUpscaleCmd()

	if cmd.Flags().Lookup("video") == nil {
		t.Error("expected flag 'video' not found")
	}
}

func TestUpscale_ShortFlags(t *testing.T) {
	cmd := newUpscaleCmd()

	flag := cmd.Flags().ShorthandLookup("v")
	if flag == nil {
		t.Error("expected short flag '-v' not found")
		return
	}
	if flag.Name != "video" {
		t.Errorf("short flag '-v' maps to '%s', expected 'video'", flag.Name)
	}
}
