package nostr

import (
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
)

// ProfileMetadata represents parsed kind 0 (profile metadata) event content
type ProfileMetadata struct {
	Name        string `json:"name"`         // Display name
	DisplayName string `json:"display_name"` // Alternative display name field
	About       string `json:"about"`        // Bio/description
	Picture     string `json:"picture"`      // Profile picture URL
	Banner      string `json:"banner"`       // Banner image URL
	Website     string `json:"website"`      // Personal website
	NIP05       string `json:"nip05"`        // NIP-05 identifier (email-like)
	LUD16       string `json:"lud16"`        // Lightning address (LNURL)
	LUD06       string `json:"lud06"`        // Lightning LNURL (deprecated)
}

// ParseProfile extracts and parses profile metadata from a kind 0 event
// Returns nil if the event is not kind 0 or if parsing fails
func ParseProfile(event *nostr.Event) *ProfileMetadata {
	if event == nil || event.Kind != 0 {
		return nil
	}

	var profile ProfileMetadata
	if err := json.Unmarshal([]byte(event.Content), &profile); err != nil {
		// Return empty profile on parse error rather than nil
		return &ProfileMetadata{}
	}

	return &profile
}

// GetDisplayName returns the best available display name
// Priority: display_name > name > empty string
func (p *ProfileMetadata) GetDisplayName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	if p.Name != "" {
		return p.Name
	}
	return ""
}

// GetLightningAddress returns the best available lightning address
// Priority: lud16 > lud06 > empty string
func (p *ProfileMetadata) GetLightningAddress() string {
	if p.LUD16 != "" {
		return p.LUD16
	}
	if p.LUD06 != "" {
		return p.LUD06
	}
	return ""
}

// HasAnyField returns true if at least one profile field is populated
func (p *ProfileMetadata) HasAnyField() bool {
	return p.Name != "" ||
		p.DisplayName != "" ||
		p.About != "" ||
		p.Picture != "" ||
		p.Banner != "" ||
		p.Website != "" ||
		p.NIP05 != "" ||
		p.LUD16 != "" ||
		p.LUD06 != ""
}
