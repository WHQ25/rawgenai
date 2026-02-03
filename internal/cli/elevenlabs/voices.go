package elevenlabs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/WHQ25/rawgenai/internal/cli/common"
	"github.com/WHQ25/rawgenai/internal/config"
	"github.com/spf13/cobra"
)

type voicesFlags struct {
	search       string
	voiceType    string
	category     string
	pageSize     int
	pageToken    string
	sort         string
	sortDir      string
	collectionID string
	voiceIDs     []string
	totalCount   bool
}

type voicesResponse struct {
	Success       bool          `json:"success"`
	Voices        []voiceItem   `json:"voices"`
	HasMore       bool          `json:"has_more"`
	TotalCount    int           `json:"total_count,omitempty"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

type voiceItem struct {
	VoiceID     string            `json:"voice_id"`
	Name        string            `json:"name"`
	Category    string            `json:"category,omitempty"`
	Description string            `json:"description,omitempty"`
	PreviewURL  string            `json:"preview_url,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func newVoiceListCmd() *cobra.Command {
	flags := &voicesFlags{}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List available voices with filtering and pagination",
		Example: `  rawgenai elevenlabs voice list
  rawgenai elevenlabs voice list --search "british"
  rawgenai elevenlabs voice list --category cloned --page-size 20
  rawgenai elevenlabs voice list --voice-type personal
  rawgenai elevenlabs voice list --sort name --sort-dir asc`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoices(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.search, "search", "", "Search term (searches name, description, labels)")
	cmd.Flags().StringVar(&flags.voiceType, "voice-type", "", "Filter by type: personal, community, default, workspace, non-default, saved")
	cmd.Flags().StringVar(&flags.category, "category", "", "Filter by category: premade, cloned, generated, professional")
	cmd.Flags().IntVar(&flags.pageSize, "page-size", 10, "Results per page (max 100)")
	cmd.Flags().StringVar(&flags.pageToken, "page-token", "", "Page token for pagination")
	cmd.Flags().StringVar(&flags.sort, "sort", "", "Sort by: created_at_unix, name")
	cmd.Flags().StringVar(&flags.sortDir, "sort-dir", "", "Sort direction: asc, desc")
	cmd.Flags().StringVar(&flags.collectionID, "collection-id", "", "Filter by collection ID")
	cmd.Flags().StringSliceVar(&flags.voiceIDs, "voice-ids", nil, "Lookup specific voice IDs (comma-separated, max 100)")
	cmd.Flags().BoolVar(&flags.totalCount, "total-count", true, "Include total count in response")

	return cmd
}

func runVoices(cmd *cobra.Command, args []string, flags *voicesFlags) error {
	// Validate page size
	if flags.pageSize < 1 || flags.pageSize > 100 {
		return common.WriteError(cmd, "invalid_page_size", "page size must be between 1 and 100")
	}

	// Validate voice type
	validVoiceTypes := map[string]bool{
		"":           true,
		"personal":   true,
		"community":  true,
		"default":    true,
		"workspace":  true,
		"non-default": true,
		"saved":      true,
	}
	if !validVoiceTypes[flags.voiceType] {
		return common.WriteError(cmd, "invalid_voice_type", "voice type must be one of: personal, community, default, workspace, non-default, saved")
	}

	// Validate category
	validCategories := map[string]bool{
		"":             true,
		"premade":      true,
		"cloned":       true,
		"generated":    true,
		"professional": true,
	}
	if !validCategories[flags.category] {
		return common.WriteError(cmd, "invalid_category", "category must be one of: premade, cloned, generated, professional")
	}

	// Validate sort
	validSorts := map[string]bool{
		"":               true,
		"created_at_unix": true,
		"name":           true,
	}
	if !validSorts[flags.sort] {
		return common.WriteError(cmd, "invalid_sort", "sort must be one of: created_at_unix, name")
	}

	// Validate sort direction
	validSortDirs := map[string]bool{
		"":     true,
		"asc":  true,
		"desc": true,
	}
	if !validSortDirs[flags.sortDir] {
		return common.WriteError(cmd, "invalid_sort_dir", "sort direction must be asc or desc")
	}

	// Validate voice IDs count
	if len(flags.voiceIDs) > 100 {
		return common.WriteError(cmd, "too_many_voice_ids", "maximum 100 voice IDs allowed")
	}

	// Check API key
	apiKey := config.GetAPIKey("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return common.WriteError(cmd, "missing_api_key", config.GetMissingKeyMessage("ELEVENLABS_API_KEY"))
	}

	// Build query parameters
	params := url.Values{}
	params.Set("page_size", strconv.Itoa(flags.pageSize))
	params.Set("include_total_count", strconv.FormatBool(flags.totalCount))

	if flags.search != "" {
		params.Set("search", flags.search)
	}
	if flags.voiceType != "" {
		params.Set("voice_type", flags.voiceType)
	}
	if flags.category != "" {
		params.Set("category", flags.category)
	}
	if flags.pageToken != "" {
		params.Set("next_page_token", flags.pageToken)
	}
	if flags.sort != "" {
		params.Set("sort", flags.sort)
	}
	if flags.sortDir != "" {
		params.Set("sort_direction", flags.sortDir)
	}
	if flags.collectionID != "" {
		params.Set("collection_id", flags.collectionID)
	}
	for _, id := range flags.voiceIDs {
		params.Add("voice_ids", id)
	}

	// Make API request (v2 endpoint)
	apiURL := fmt.Sprintf("https://api.elevenlabs.io/v2/voices?%s", params.Encode())
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return common.WriteError(cmd, "internal_error", fmt.Sprintf("cannot create request: %s", err.Error()))
	}

	req.Header.Set("xi-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return handleHTTPError(cmd, err)
	}
	defer resp.Body.Close()

	// Handle API errors
	if resp.StatusCode != http.StatusOK {
		return handleAPIErrorResponse(cmd, resp)
	}

	// Parse response
	var apiResp struct {
		Voices []struct {
			VoiceID     string            `json:"voice_id"`
			Name        string            `json:"name"`
			Category    string            `json:"category"`
			Description string            `json:"description"`
			PreviewURL  string            `json:"preview_url"`
			Labels      map[string]string `json:"labels"`
		} `json:"voices"`
		HasMore       bool   `json:"has_more"`
		TotalCount    int    `json:"total_count"`
		NextPageToken string `json:"next_page_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return common.WriteError(cmd, "invalid_response", fmt.Sprintf("cannot parse response: %s", err.Error()))
	}

	// Build result
	voices := make([]voiceItem, len(apiResp.Voices))
	for i, v := range apiResp.Voices {
		voices[i] = voiceItem{
			VoiceID:     v.VoiceID,
			Name:        v.Name,
			Category:    v.Category,
			Description: v.Description,
			PreviewURL:  v.PreviewURL,
			Labels:      v.Labels,
		}
	}

	result := voicesResponse{
		Success:       true,
		Voices:        voices,
		HasMore:       apiResp.HasMore,
		TotalCount:    apiResp.TotalCount,
		NextPageToken: apiResp.NextPageToken,
	}

	return common.WriteSuccess(cmd, result)
}
