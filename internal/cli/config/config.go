package config

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

// Cmd is the config command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage rawgenai configuration",
	Long:  "Set, unset, and list API keys stored in ~/.config/rawgenai/config.json",
}

func init() {
	Cmd.AddCommand(setCmd)
	Cmd.AddCommand(unsetCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(pathCmd)
}

// Response types
type successResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
}

type listResponse struct {
	Success bool              `json:"success"`
	Keys    map[string]string `json:"keys"`
}

type errorResponse struct {
	Success bool       `json:"success"`
	Error   *errorInfo `json:"error"`
}

type errorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeSuccess(cmd *cobra.Command, msg string) {
	resp := successResponse{Success: true, Message: msg}
	output, _ := json.Marshal(resp)
	fmt.Fprintln(cmd.OutOrStdout(), string(output))
}

func writeError(cmd *cobra.Command, code, msg string) {
	resp := errorResponse{
		Success: false,
		Error:   &errorInfo{Code: code, Message: msg},
	}
	output, _ := json.Marshal(resp)
	fmt.Fprintln(cmd.ErrOrStderr(), string(output))
}

// set command
var setCmd = &cobra.Command{
	Use:           "set <key> <value>",
	Short:         "Set a config value",
	Long:          "Set an API key in the config file. Accepts both formats: openai_api_key or OPENAI_API_KEY",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := config.NormalizeKey(args[0])
		if key == "" {
			writeError(cmd, "invalid_key", fmt.Sprintf("unknown key: %s. Valid keys: %v", args[0], config.ValidKeys()))
			return fmt.Errorf("invalid_key")
		}

		cfg, err := config.Load()
		if err != nil {
			writeError(cmd, "load_error", fmt.Sprintf("failed to load config: %s", err.Error()))
			return err
		}

		if err := cfg.Set(key, args[1]); err != nil {
			writeError(cmd, "set_error", err.Error())
			return err
		}

		if err := config.Save(cfg); err != nil {
			writeError(cmd, "save_error", fmt.Sprintf("failed to save config: %s", err.Error()))
			return err
		}

		writeSuccess(cmd, fmt.Sprintf("Set %s", key))
		return nil
	},
}

// unset command
var unsetCmd = &cobra.Command{
	Use:           "unset <key>",
	Short:         "Remove a config value",
	Long:          "Remove an API key from the config file.",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := config.NormalizeKey(args[0])
		if key == "" {
			writeError(cmd, "invalid_key", fmt.Sprintf("unknown key: %s. Valid keys: %v", args[0], config.ValidKeys()))
			return fmt.Errorf("invalid_key")
		}

		cfg, err := config.Load()
		if err != nil {
			writeError(cmd, "load_error", fmt.Sprintf("failed to load config: %s", err.Error()))
			return err
		}

		if err := cfg.Unset(key); err != nil {
			writeError(cmd, "unset_error", err.Error())
			return err
		}

		if err := config.Save(cfg); err != nil {
			writeError(cmd, "save_error", fmt.Sprintf("failed to save config: %s", err.Error()))
			return err
		}

		writeSuccess(cmd, fmt.Sprintf("Unset %s", key))
		return nil
	},
}

// list command
var listCmd = &cobra.Command{
	Use:           "list",
	Short:         "List all config values",
	Long:          "List all configured API keys with masked values.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			writeError(cmd, "load_error", fmt.Sprintf("failed to load config: %s", err.Error()))
			return err
		}

		keys := cfg.List()

		// Sort keys for consistent output
		sortedKeys := make([]string, 0, len(keys))
		for k := range keys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		sortedMap := make(map[string]string)
		for _, k := range sortedKeys {
			sortedMap[k] = keys[k]
		}

		resp := listResponse{Success: true, Keys: sortedMap}
		output, _ := json.Marshal(resp)
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
		return nil
	},
}

// path command
var pathCmd = &cobra.Command{
	Use:           "path",
	Short:         "Show config file path",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp := successResponse{Success: true, Path: config.Path()}
		output, _ := json.Marshal(resp)
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
		return nil
	},
}
