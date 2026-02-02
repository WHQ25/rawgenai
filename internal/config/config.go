package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds API keys and other configuration
type Config struct {
	OpenAIAPIKey     string `json:"openai_api_key,omitempty"`
	GeminiAPIKey     string `json:"gemini_api_key,omitempty"`
	GoogleAPIKey     string `json:"google_api_key,omitempty"`
	ElevenLabsAPIKey string `json:"elevenlabs_api_key,omitempty"`
	XAIAPIKey        string `json:"xai_api_key,omitempty"`
	ArkAPIKey        string `json:"ark_api_key,omitempty"`
	SeedAppID        string `json:"seed_app_id,omitempty"`
	SeedAccessToken  string `json:"seed_access_token,omitempty"`
}

// validKeys maps normalized key names to their JSON field names
var validKeys = map[string]string{
	"openai_api_key":     "openai_api_key",
	"gemini_api_key":     "gemini_api_key",
	"google_api_key":     "google_api_key",
	"elevenlabs_api_key": "elevenlabs_api_key",
	"xai_api_key":        "xai_api_key",
	"ark_api_key":        "ark_api_key",
	"seed_app_id":        "seed_app_id",
	"seed_access_token":  "seed_access_token",
}

// envToConfigKey maps environment variable names to config keys
var envToConfigKey = map[string]string{
	"OPENAI_API_KEY":     "openai_api_key",
	"GEMINI_API_KEY":     "gemini_api_key",
	"GOOGLE_API_KEY":     "google_api_key",
	"ELEVENLABS_API_KEY": "elevenlabs_api_key",
	"XAI_API_KEY":        "xai_api_key",
	"ARK_API_KEY":        "ark_api_key",
	"SEED_APP_ID":        "seed_app_id",
	"SEED_ACCESS_TOKEN":  "seed_access_token",
}

// Path returns the config file path
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "rawgenai", "config.json")
}

// Load loads config from file
func Load() (*Config, error) {
	path := Path()
	if path == "" {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save saves config to file
func Save(cfg *Config) error {
	path := Path()
	if path == "" {
		return fmt.Errorf("cannot determine config path")
	}

	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// NormalizeKey converts env var format to config key format
// e.g., "OPENAI_API_KEY" -> "openai_api_key"
func NormalizeKey(key string) string {
	lower := strings.ToLower(key)
	if _, ok := validKeys[lower]; ok {
		return lower
	}
	return ""
}

// GetAPIKey returns the API key for the given environment variable name(s)
// Tries each key in order, returns the first non-empty value
// Priority for each key: environment variable > config file
func GetAPIKey(envNames ...string) string {
	cfg, _ := Load()

	for _, envName := range envNames {
		// Check environment variable first
		if val := os.Getenv(envName); val != "" {
			return val
		}

		// Check config file
		if cfg != nil {
			if configKey, ok := envToConfigKey[envName]; ok {
				if val := cfg.Get(configKey); val != "" {
					return val
				}
			}
		}
	}

	return ""
}

// GetMissingKeyMessage returns a helpful error message for missing API key
func GetMissingKeyMessage(envNames ...string) string {
	if len(envNames) == 0 {
		return "API key not found"
	}
	if len(envNames) == 1 {
		envName := envNames[0]
		configKey := envToConfigKey[envName]
		if configKey == "" {
			configKey = strings.ToLower(envName)
		}
		return fmt.Sprintf("%s not found. Set it with: rawgenai config set %s <your-key>", envName, configKey)
	}
	// Multiple keys (e.g., GEMINI_API_KEY or GOOGLE_API_KEY)
	configKey := envToConfigKey[envNames[0]]
	if configKey == "" {
		configKey = strings.ToLower(envNames[0])
	}
	return fmt.Sprintf("%s not found. Set it with: rawgenai config set %s <your-key>", strings.Join(envNames, " or "), configKey)
}

// Get returns a config value by key
func (c *Config) Get(key string) string {
	switch key {
	case "openai_api_key":
		return c.OpenAIAPIKey
	case "gemini_api_key":
		return c.GeminiAPIKey
	case "google_api_key":
		return c.GoogleAPIKey
	case "elevenlabs_api_key":
		return c.ElevenLabsAPIKey
	case "xai_api_key":
		return c.XAIAPIKey
	case "ark_api_key":
		return c.ArkAPIKey
	case "seed_app_id":
		return c.SeedAppID
	case "seed_access_token":
		return c.SeedAccessToken
	default:
		return ""
	}
}

// Set sets a config value by key
func (c *Config) Set(key, value string) error {
	switch key {
	case "openai_api_key":
		c.OpenAIAPIKey = value
	case "gemini_api_key":
		c.GeminiAPIKey = value
	case "google_api_key":
		c.GoogleAPIKey = value
	case "elevenlabs_api_key":
		c.ElevenLabsAPIKey = value
	case "xai_api_key":
		c.XAIAPIKey = value
	case "ark_api_key":
		c.ArkAPIKey = value
	case "seed_app_id":
		c.SeedAppID = value
	case "seed_access_token":
		c.SeedAccessToken = value
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return nil
}

// Unset removes a config value by key
func (c *Config) Unset(key string) error {
	return c.Set(key, "")
}

// List returns all keys with masked values
func (c *Config) List() map[string]string {
	result := make(map[string]string)
	for key := range validKeys {
		val := c.Get(key)
		if val == "" {
			result[key] = "(not set)"
		} else {
			result[key] = maskValue(val)
		}
	}
	return result
}

// maskValue masks the middle of a value, showing only first 3 and last 3 chars
func maskValue(val string) string {
	if len(val) <= 8 {
		return "***"
	}
	return val[:3] + "***" + val[len(val)-3:]
}

// ValidKeys returns the list of valid config keys
func ValidKeys() []string {
	keys := make([]string, 0, len(validKeys))
	for k := range validKeys {
		keys = append(keys, k)
	}
	return keys
}
