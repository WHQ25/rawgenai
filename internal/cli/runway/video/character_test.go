package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCharacter_MissingCharacter(t *testing.T) {
	refFile := createTempFile(t, "ref.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newCharacterCmd())
	_, stderr, err := executeCommand(cmd, "character", "-r", refFile)

	if err == nil {
		t.Fatal("expected error for missing character")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_character" {
		t.Errorf("expected error code 'missing_character', got: %s", errorObj["code"])
	}
}

func TestCharacter_MissingReference(t *testing.T) {
	charFile := createTempFile(t, "char.jpg", "fake image")

	cmd := newTestCmd()
	cmd.AddCommand(newCharacterCmd())
	_, stderr, err := executeCommand(cmd, "character", "-c", charFile)

	if err == nil {
		t.Fatal("expected error for missing reference")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_reference" {
		t.Errorf("expected error code 'missing_reference', got: %s", errorObj["code"])
	}
}

func TestCharacter_InvalidExpression(t *testing.T) {
	charFile := createTempFile(t, "char.jpg", "fake image")
	refFile := createTempFile(t, "ref.mp4", "fake video")

	tests := []string{"0", "6", "-1"}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			cmd := newTestCmd()
			cmd.AddCommand(newCharacterCmd())
			_, stderr, err := executeCommand(cmd, "character", "-c", charFile, "-r", refFile, "-e", expr)

			if err == nil {
				t.Fatal("expected error for invalid expression")
			}

			var resp map[string]any
			json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

			errorObj := resp["error"].(map[string]any)
			if errorObj["code"] != "invalid_expression" {
				t.Errorf("expected error code 'invalid_expression', got: %s", errorObj["code"])
			}
		})
	}
}

func TestCharacter_CharacterNotFound(t *testing.T) {
	refFile := createTempFile(t, "ref.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newCharacterCmd())
	_, stderr, err := executeCommand(cmd, "character", "-c", "/nonexistent/char.jpg", "-r", refFile)

	if err == nil {
		t.Fatal("expected error for character not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "character_not_found" {
		t.Errorf("expected error code 'character_not_found', got: %s", errorObj["code"])
	}
}

func TestCharacter_ReferenceNotFound(t *testing.T) {
	charFile := createTempFile(t, "char.jpg", "fake image")

	cmd := newTestCmd()
	cmd.AddCommand(newCharacterCmd())
	_, stderr, err := executeCommand(cmd, "character", "-c", charFile, "-r", "/nonexistent/ref.mp4")

	if err == nil {
		t.Fatal("expected error for reference not found")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "reference_not_found" {
		t.Errorf("expected error code 'reference_not_found', got: %s", errorObj["code"])
	}
}

func TestCharacter_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)
	charFile := createTempFile(t, "char.jpg", "fake image")
	refFile := createTempFile(t, "ref.mp4", "fake video")

	cmd := newTestCmd()
	cmd.AddCommand(newCharacterCmd())
	_, stderr, err := executeCommand(cmd, "character", "-c", charFile, "-r", refFile)

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

func TestCharacter_AllFlags(t *testing.T) {
	cmd := newCharacterCmd()

	expectedFlags := []string{"character", "character-type", "reference", "seed", "body-control", "expression", "ratio", "public-figure"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag '%s' not found", flag)
		}
	}
}

func TestCharacter_DefaultValues(t *testing.T) {
	cmd := newCharacterCmd()

	defaults := map[string]string{
		"character-type": "image",
		"expression":     "3",
		"ratio":          "1280:720",
		"public-figure":  "auto",
		"body-control":   "false",
	}

	for flag, expected := range defaults {
		f := cmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("flag '%s' not found", flag)
			continue
		}
		if f.DefValue != expected {
			t.Errorf("flag '%s' default is '%s', expected '%s'", flag, f.DefValue, expected)
		}
	}
}
