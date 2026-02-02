package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"OPENAI_API_KEY", "openai_api_key"},
		{"openai_api_key", "openai_api_key"},
		{"GEMINI_API_KEY", "gemini_api_key"},
		{"ARK_API_KEY", "ark_api_key"},
		{"SEED_APP_ID", "seed_app_id"},
		{"SEED_ACCESS_TOKEN", "seed_access_token"},
		{"KLING_ACCESS_KEY", "kling_access_key"},
		{"KLING_SECRET_KEY", "kling_secret_key"},
		{"KLING_BASE_URL", "kling_base_url"},
		{"INVALID_KEY", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeKey(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAPIKey_EnvPriority(t *testing.T) {
	// Env var should take priority over config
	t.Setenv("OPENAI_API_KEY", "env-key")

	result := GetAPIKey("OPENAI_API_KEY")
	if result != "env-key" {
		t.Errorf("GetAPIKey should return env var, got: %s", result)
	}
}

func TestGetAPIKey_EmptyEnv(t *testing.T) {
	// Use temp HOME to avoid reading real config file
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("OPENAI_API_KEY", "")

	// Without config file, should return empty
	result := GetAPIKey("OPENAI_API_KEY")
	if result != "" {
		t.Errorf("GetAPIKey should return empty when no env and no config, got: %s", result)
	}
}

func TestGetAPIKey_MultipleKeys(t *testing.T) {
	// Use temp HOME to avoid reading real config file
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "google-key")

	result := GetAPIKey("GEMINI_API_KEY", "GOOGLE_API_KEY")
	if result != "google-key" {
		t.Errorf("GetAPIKey should fallback to second key, got: %s", result)
	}
}

func TestGetAPIKey_FirstKeyFound(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("GOOGLE_API_KEY", "google-key")

	result := GetAPIKey("GEMINI_API_KEY", "GOOGLE_API_KEY")
	if result != "gemini-key" {
		t.Errorf("GetAPIKey should return first found key, got: %s", result)
	}
}

func TestGetMissingKeyMessage_Single(t *testing.T) {
	msg := GetMissingKeyMessage("OPENAI_API_KEY")
	expected := "OPENAI_API_KEY not found. Set it with: rawgenai config set openai_api_key <your-key>"
	if msg != expected {
		t.Errorf("GetMissingKeyMessage = %q, want %q", msg, expected)
	}
}

func TestGetMissingKeyMessage_Multiple(t *testing.T) {
	msg := GetMissingKeyMessage("GEMINI_API_KEY", "GOOGLE_API_KEY")
	expected := "GEMINI_API_KEY or GOOGLE_API_KEY not found. Set it with: rawgenai config set gemini_api_key <your-key>"
	if msg != expected {
		t.Errorf("GetMissingKeyMessage = %q, want %q", msg, expected)
	}
}

func TestConfig_GetSet(t *testing.T) {
	cfg := &Config{}

	tests := []struct {
		key   string
		value string
	}{
		{"openai_api_key", "sk-test"},
		{"gemini_api_key", "gemini-test"},
		{"google_api_key", "google-test"},
		{"elevenlabs_api_key", "el-test"},
		{"xai_api_key", "xai-test"},
		{"ark_api_key", "ark-test"},
		{"seed_app_id", "app-test"},
		{"seed_access_token", "token-test"},
		{"kling_access_key", "kling-access-test"},
		{"kling_secret_key", "kling-secret-test"},
		{"kling_base_url", "https://api.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := cfg.Set(tt.key, tt.value)
			if err != nil {
				t.Fatalf("Set(%q) error: %v", tt.key, err)
			}

			got := cfg.Get(tt.key)
			if got != tt.value {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

func TestConfig_SetInvalidKey(t *testing.T) {
	cfg := &Config{}
	err := cfg.Set("invalid_key", "value")
	if err == nil {
		t.Error("Set with invalid key should return error")
	}
}

func TestConfig_Unset(t *testing.T) {
	cfg := &Config{OpenAIAPIKey: "test-key"}

	err := cfg.Unset("openai_api_key")
	if err != nil {
		t.Fatalf("Unset error: %v", err)
	}

	if cfg.OpenAIAPIKey != "" {
		t.Error("Unset should clear the value")
	}
}

func TestConfig_List(t *testing.T) {
	cfg := &Config{
		OpenAIAPIKey: "sk-test123abc",
		GeminiAPIKey: "",
	}

	list := cfg.List()

	// Check masked value
	if list["openai_api_key"] != "sk-***abc" {
		t.Errorf("List should mask value, got: %s", list["openai_api_key"])
	}

	// Check not set
	if list["gemini_api_key"] != "(not set)" {
		t.Errorf("List should show (not set), got: %s", list["gemini_api_key"])
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-test123abc", "sk-***abc"},
		{"short", "***"},
		{"12345678", "***"},
		{"123456789", "123***789"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskValue(tt.input)
			if result != tt.expected {
				t.Errorf("maskValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidKeys(t *testing.T) {
	keys := ValidKeys()

	expectedKeys := []string{
		"openai_api_key", "gemini_api_key", "google_api_key",
		"elevenlabs_api_key", "xai_api_key", "ark_api_key",
		"seed_app_id", "seed_access_token",
		"kling_access_key", "kling_secret_key", "kling_base_url",
		"runway_api_key",
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("ValidKeys() returned %d keys, want %d", len(keys), len(expectedKeys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	for _, expected := range expectedKeys {
		if !keySet[expected] {
			t.Errorf("ValidKeys() missing key: %s", expected)
		}
	}
}

func TestPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Error("Path() should not return empty string")
	}

	// Should contain config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("Path() should end with config.json, got: %s", path)
	}

	// Should be in .config/rawgenai
	dir := filepath.Dir(path)
	if filepath.Base(dir) != "rawgenai" {
		t.Errorf("Path() should be in rawgenai dir, got: %s", dir)
	}
}

func TestLoadSave(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home for test
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Test save
	cfg := &Config{
		OpenAIAPIKey: "test-key",
		SeedAppID:    "app-id",
	}

	err = Save(cfg)
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Test load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if loaded.OpenAIAPIKey != "test-key" {
		t.Errorf("Loaded OpenAIAPIKey = %q, want %q", loaded.OpenAIAPIKey, "test-key")
	}

	if loaded.SeedAppID != "app-id" {
		t.Errorf("Loaded SeedAppID = %q, want %q", loaded.SeedAppID, "app-id")
	}
}

func TestLoad_NoFile(t *testing.T) {
	// Create temp dir without config file
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load should return empty config, not error
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error when file doesn't exist: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load should return empty config, not nil")
	}

	if cfg.OpenAIAPIKey != "" {
		t.Error("Empty config should have empty values")
	}
}
