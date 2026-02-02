package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

var validTags = map[string]string{
	"o_101": "热梗",
	"o_102": "人物",
	"o_103": "动物",
	"o_104": "道具",
	"o_105": "服饰",
	"o_106": "场景",
	"o_107": "特效",
	"o_108": "其他",
}

type elementCreateFlags struct {
	description  string
	frontalImage string
	refImages    []string
	tags         []string
}

func newElementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "element",
		Short: "Manage video generation elements",
		Long:  "Create and list custom elements for video generation.",
	}

	cmd.AddCommand(newElementCreateCmd())
	cmd.AddCommand(newElementListCmd())
	cmd.AddCommand(newElementDeleteCmd())

	return cmd
}

func newElementCreateCmd() *cobra.Command {
	flags := &elementCreateFlags{}

	cmd := &cobra.Command{
		Use:           "create <name>",
		Short:         "Create a custom element",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runElementCreate(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.description, "description", "d", "", "Element description (max 100 chars)")
	cmd.Flags().StringVarP(&flags.frontalImage, "frontal", "f", "", "Frontal reference image path (required)")
	cmd.Flags().StringArrayVarP(&flags.refImages, "ref", "r", nil, "Additional reference images (1-3)")
	cmd.Flags().StringArrayVarP(&flags.tags, "tag", "t", nil, "Tags: o_101(热梗), o_102(人物), o_103(动物), o_104(道具), o_105(服饰), o_106(场景), o_107(特效), o_108(其他)")

	return cmd
}

func runElementCreate(cmd *cobra.Command, args []string, flags *elementCreateFlags) error {
	// Validate name
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_name", "element name is required")
	}
	name := args[0]

	if len(name) > 20 {
		return common.WriteError(cmd, "invalid_name", "element name must be at most 20 characters")
	}

	// Validate description
	if flags.description == "" {
		return common.WriteError(cmd, "missing_description", "element description is required (-d)")
	}
	if len(flags.description) > 100 {
		return common.WriteError(cmd, "invalid_description", "element description must be at most 100 characters")
	}

	// Validate frontal image
	if flags.frontalImage == "" {
		return common.WriteError(cmd, "missing_frontal", "frontal image is required (-f)")
	}
	if !isURL(flags.frontalImage) {
		if _, err := os.Stat(flags.frontalImage); os.IsNotExist(err) {
			return common.WriteError(cmd, "frontal_not_found", fmt.Sprintf("frontal image not found: %s", flags.frontalImage))
		}
	}

	// Validate ref images
	if len(flags.refImages) < 1 || len(flags.refImages) > 3 {
		return common.WriteError(cmd, "invalid_ref_count", "must provide 1-3 reference images (-r)")
	}
	for _, img := range flags.refImages {
		if !isURL(img) {
			if _, err := os.Stat(img); os.IsNotExist(err) {
				return common.WriteError(cmd, "ref_not_found", fmt.Sprintf("reference image not found: %s", img))
			}
		}
	}

	// Validate tags
	for _, tag := range flags.tags {
		if _, ok := validTags[tag]; !ok {
			return common.WriteError(cmd, "invalid_tag", fmt.Sprintf("invalid tag: %s", tag))
		}
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := generateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Resolve frontal image URL
	frontalURL, err := resolveImageURL(flags.frontalImage)
	if err != nil {
		return common.WriteError(cmd, "frontal_read_error", fmt.Sprintf("cannot read frontal image: %s", err.Error()))
	}

	// Resolve ref images
	refList := []map[string]string{}
	for _, img := range flags.refImages {
		imgURL, err := resolveImageURL(img)
		if err != nil {
			return common.WriteError(cmd, "ref_read_error", fmt.Sprintf("cannot read reference image: %s", err.Error()))
		}
		refList = append(refList, map[string]string{"image_url": imgURL})
	}

	// Build request body
	body := map[string]any{
		"element_name":          name,
		"element_description":   flags.description,
		"element_frontal_image": frontalURL,
		"element_refer_list":    refList,
	}

	if len(flags.tags) > 0 {
		tagList := []map[string]string{}
		for _, tag := range flags.tags {
			tagList = append(tagList, map[string]string{"tag_id": tag})
		}
		body["tag_list"] = tagList
	}

	// Serialize request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/general/custom-elements", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			ElementID   int64  `json:"element_id"`
			ElementName string `json:"element_name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	if result.Data == nil {
		return common.WriteError(cmd, "response_error", "no data in response")
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":      true,
		"element_id":   result.Data.ElementID,
		"element_name": result.Data.ElementName,
	})
}

type elementListFlags struct {
	elementType string
	limit       int
	page        int
}

func newElementListCmd() *cobra.Command {
	flags := &elementListFlags{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List elements",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runElementList(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.elementType, "type", "t", "custom", "Element type: custom, official")
	cmd.Flags().IntVarP(&flags.limit, "limit", "l", 30, "Maximum elements to return (1-500)")
	cmd.Flags().IntVarP(&flags.page, "page", "p", 1, "Page number")

	return cmd
}

func runElementList(cmd *cobra.Command, flags *elementListFlags) error {
	// Validate type
	if flags.elementType != "custom" && flags.elementType != "official" {
		return common.WriteError(cmd, "invalid_type", "type must be 'custom' or 'official'")
	}

	// Validate limit
	if flags.limit < 1 || flags.limit > 500 {
		return common.WriteError(cmd, "invalid_limit", "limit must be between 1 and 500")
	}

	// Validate page
	if flags.page < 1 {
		return common.WriteError(cmd, "invalid_page", "page must be at least 1")
	}

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := generateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Determine endpoint
	endpoint := "/v1/general/custom-elements"
	if flags.elementType == "official" {
		endpoint = "/v1/general/presets-elements"
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s?pageNum=%d&pageSize=%d", getKlingAPIBase(), endpoint, flags.page, flags.limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			ElementID          int64  `json:"element_id"`
			ElementName        string `json:"element_name"`
			ElementDescription string `json:"element_description"`
			ElementFrontalImg  string `json:"element_frontal_image"`
			OwnedBy            string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	// Build response
	elements := []map[string]any{}
	if result.Data != nil {
		for _, e := range result.Data {
			elements = append(elements, map[string]any{
				"element_id":   e.ElementID,
				"name":         e.ElementName,
				"description":  e.ElementDescription,
				"frontal_url":  e.ElementFrontalImg,
				"owned_by":     e.OwnedBy,
			})
		}
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":  true,
		"type":     flags.elementType,
		"elements": elements,
		"count":    len(elements),
	})
}

type elementDeleteFlags struct{}

func newElementDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete <element_id>",
		Short:         "Delete a custom element",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runElementDelete(cmd, args)
		},
	}

	return cmd
}

func runElementDelete(cmd *cobra.Command, args []string) error {
	// Validate element ID
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return common.WriteError(cmd, "missing_element_id", "element ID is required")
	}
	elementID := args[0]

	// Check API keys
	accessKey := config.GetAPIKey("KLING_ACCESS_KEY")
	secretKey := config.GetAPIKey("KLING_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("KLING_ACCESS_KEY")+" and "+config.GetMissingKeyMessage("KLING_SECRET_KEY"))
	}

	// Generate JWT token
	token, err := generateJWT(accessKey, secretKey)
	if err != nil {
		return common.WriteError(cmd, "auth_error", fmt.Sprintf("failed to generate JWT: %s", err.Error()))
	}

	// Build request body
	body := map[string]any{
		"element_id": elementID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot serialize request: %s", err.Error()))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", getKlingAPIBase()+"/v1/general/delete-elements", bytes.NewReader(jsonBody))
	if err != nil {
		return common.WriteError(cmd, "request_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return handleAPIError(cmd, err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot read response: %s", err.Error()))
	}

	// Parse response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			TaskID     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return common.WriteError(cmd, "response_error", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	if result.Code != 0 {
		return handleKlingError(cmd, result.Code, result.Message)
	}

	return common.WriteSuccess(cmd, map[string]any{
		"success":    true,
		"element_id": elementID,
		"status":     "deleted",
	})
}
