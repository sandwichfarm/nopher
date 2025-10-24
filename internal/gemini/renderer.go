package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nophr/internal/aggregates"
	"github.com/sandwich/nophr/internal/config"
	"github.com/sandwich/nophr/internal/entities"
	"github.com/sandwich/nophr/internal/markdown"
	nostrclient "github.com/sandwich/nophr/internal/nostr"
	"github.com/sandwich/nophr/internal/presentation"
	"github.com/sandwich/nophr/internal/storage"
)

// Renderer renders Nostr events as Gemtext
type Renderer struct {
	parser   *markdown.Parser
	config   *config.Config
	loader   *presentation.Loader
	resolver *entities.Resolver
}

// NewRenderer creates a new event renderer
func NewRenderer(cfg *config.Config, st *storage.Storage) *Renderer {
	return &Renderer{
		parser:   markdown.NewParser(),
		config:   cfg,
		loader:   presentation.NewLoader(cfg),
		resolver: entities.NewResolver(st),
	}
}

// RenderHome renders the home page
func (r *Renderer) RenderHome() string {
	var sb strings.Builder

	sb.WriteString("# nophr - Nostr Gateway\n\n")
	sb.WriteString("Browse Nostr content via Gemini protocol\n\n")
	sb.WriteString("## Navigation\n\n")
	sb.WriteString("=> /notes Notes\n")
	sb.WriteString("=> /articles Articles\n")
	sb.WriteString("=> /replies Replies\n")
	sb.WriteString("=> /mentions Mentions\n")
	sb.WriteString("=> /search Search\n")
	sb.WriteString("=> /diagnostics Diagnostics\n")
	sb.WriteString("\n")
	sb.WriteString("Powered by nophr\n")

	return r.applyHeadersFooters(sb.String(), "home")
}

// RenderNote renders a note event as gemtext
func (r *Renderer) RenderNote(event *nostr.Event, agg *aggregates.EventAggregates, threadURL, homeURL string) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Note by %s\n", truncatePubkey(event.PubKey)))
	sb.WriteString(fmt.Sprintf("Posted: %s\n\n", formatTimestamp(event.CreatedAt)))

	// Content (resolve NIP-19 entities, then render markdown as gemtext)
	content := event.Content
	ctx := context.Background()
	content = r.resolver.ReplaceEntities(ctx, content, entities.PlainTextFormatter)

	rendered, _ := r.parser.RenderGemini([]byte(content), nil)
	sb.WriteString(rendered)
	sb.WriteString("\n")

	// Aggregates
	if agg != nil && agg.HasInteractions() {
		sb.WriteString("## Interactions\n\n")
		sb.WriteString(r.renderAggregates(agg))
		sb.WriteString("\n")
	}

	// Navigation
	sb.WriteString("## Actions\n\n")
	sb.WriteString(fmt.Sprintf("=> %s View Thread\n", threadURL))
	sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))

	return sb.String()
}

// RenderProfile renders a profile event
func (r *Renderer) RenderProfile(profileEvent *nostr.Event, homeURL string) string {
	var sb strings.Builder

	// Parse profile metadata
	profile := nostrclient.ParseProfile(profileEvent)
	if profile == nil {
		// Fallback for invalid profile
		sb.WriteString(fmt.Sprintf("# Profile: %s\n\n", truncatePubkey(profileEvent.PubKey)))
		sb.WriteString("Invalid profile data\n\n")
		sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))
		return sb.String()
	}

	// Header with display name
	displayName := profile.GetDisplayName()
	if displayName == "" {
		displayName = truncatePubkey(profileEvent.PubKey)
	}

	sb.WriteString(fmt.Sprintf("# %s\n\n", displayName))

	// Pubkey
	sb.WriteString(fmt.Sprintf("Pubkey: %s\n\n", profileEvent.PubKey))

	// Name fields
	if profile.Name != "" {
		sb.WriteString(fmt.Sprintf("**Name:** %s\n", profile.Name))
	}
	if profile.DisplayName != "" && profile.DisplayName != profile.Name {
		sb.WriteString(fmt.Sprintf("**Display Name:** %s\n", profile.DisplayName))
	}
	if profile.Name != "" || profile.DisplayName != "" {
		sb.WriteString("\n")
	}

	// About/Bio
	if profile.About != "" {
		sb.WriteString("## About\n\n")
		sb.WriteString(profile.About)
		sb.WriteString("\n\n")
	}

	// Contact information section
	hasContact := profile.Website != "" || profile.NIP05 != "" || profile.GetLightningAddress() != ""
	if hasContact {
		sb.WriteString("## Contact & Links\n\n")
		if profile.Website != "" {
			sb.WriteString(fmt.Sprintf("=> %s Website\n", profile.Website))
		}
		if profile.NIP05 != "" {
			sb.WriteString(fmt.Sprintf("**NIP-05:** %s\n", profile.NIP05))
		}
		lightningAddr := profile.GetLightningAddress()
		if lightningAddr != "" {
			sb.WriteString(fmt.Sprintf("**Lightning:** %s\n", lightningAddr))
		}
		sb.WriteString("\n")
	}

	// Media section
	hasMedia := profile.Picture != "" || profile.Banner != ""
	if hasMedia {
		sb.WriteString("## Media\n\n")
		if profile.Picture != "" {
			sb.WriteString(fmt.Sprintf("=> %s Profile Picture\n", profile.Picture))
		}
		if profile.Banner != "" {
			sb.WriteString(fmt.Sprintf("=> %s Banner Image\n", profile.Banner))
		}
		sb.WriteString("\n")
	}

	// Navigation
	sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))

	return sb.String()
}

// RenderThread renders a thread with replies
func (r *Renderer) RenderThread(root *aggregates.EnrichedEvent, replies []*aggregates.EnrichedEvent, homeURL string) string {
	var sb strings.Builder

	sb.WriteString("# Thread\n\n")

	// Root post
	sb.WriteString("## Root Post\n\n")
	sb.WriteString(fmt.Sprintf("By %s - %s\n\n", truncatePubkey(root.Event.PubKey), formatTimestamp(root.Event.CreatedAt)))

	// Render content
	content, _ := r.parser.RenderGemini([]byte(root.Event.Content), nil)
	sb.WriteString(content)
	sb.WriteString("\n")

	// Root aggregates
	if root.Aggregates != nil && root.Aggregates.HasInteractions() {
		sb.WriteString(r.renderAggregates(root.Aggregates))
		sb.WriteString("\n")
	}

	// Replies
	if len(replies) > 0 {
		sb.WriteString(fmt.Sprintf("## Replies (%d)\n\n", len(replies)))

		for i, reply := range replies {
			sb.WriteString(fmt.Sprintf("### Reply %d\n\n", i+1))
			sb.WriteString(fmt.Sprintf("By %s - %s\n\n", truncatePubkey(reply.Event.PubKey), formatTimestamp(reply.Event.CreatedAt)))

			// Reply content
			replyContent, _ := r.parser.RenderGemini([]byte(reply.Event.Content), nil)
			sb.WriteString(replyContent)
			sb.WriteString("\n")

			// Reply link
			sb.WriteString(fmt.Sprintf("=> /note/%s View Reply\n\n", reply.Event.ID))
		}
	} else {
		sb.WriteString("## Replies\n\nNo replies yet.\n\n")
	}

	sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))

	return sb.String()
}

// RenderNoteList renders a list of notes with summaries
func (r *Renderer) RenderNoteList(notes []*aggregates.EnrichedEvent, title, homeURL string) string {
	var sb strings.Builder

	// Determine page name from title for headers/footers
	// Map common titles to page names
	pageName := "notes" // default
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, "article") {
		pageName = "articles"
	} else if strings.Contains(titleLower, "repl") {
		pageName = "replies"
	} else if strings.Contains(titleLower, "mention") {
		pageName = "mentions"
	}

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(notes) == 0 {
		sb.WriteString("No notes yet.\n\n")
		sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))
		return r.applyHeadersFooters(sb.String(), pageName)
	}

	for i, note := range notes {
		// Extract first line of content as summary
		content := note.Event.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		firstLine := strings.Split(content, "\n")[0]

		sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, firstLine))
		sb.WriteString(fmt.Sprintf("By %s - %s\n", truncatePubkey(note.Event.PubKey), formatTimestamp(note.Event.CreatedAt)))

		if note.Aggregates != nil && note.Aggregates.HasInteractions() {
			sb.WriteString(r.renderAggregates(note.Aggregates))
		}

		sb.WriteString(fmt.Sprintf("\n=> /note/%s Read Full Note\n\n", note.Event.ID))
	}

	sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))

	return r.applyHeadersFooters(sb.String(), pageName)
}

// renderAggregates renders interaction stats (for feed view)
func (r *Renderer) renderAggregates(agg *aggregates.EventAggregates) string {
	if !r.config.Display.Feed.ShowInteractions {
		return ""
	}
	return r.buildAggregatesString(agg, r.config.Display.Feed.ShowReplies, r.config.Display.Feed.ShowReactions, r.config.Display.Feed.ShowZaps)
}

// renderAggregatesForDetail renders interaction stats for detail view
func (r *Renderer) renderAggregatesForDetail(agg *aggregates.EventAggregates) string {
	return r.buildAggregatesString(agg, r.config.Display.Detail.ShowReplies, r.config.Display.Detail.ShowReactions, r.config.Display.Detail.ShowZaps)
}

// buildAggregatesString builds the aggregates string based on what should be shown
func (r *Renderer) buildAggregatesString(agg *aggregates.EventAggregates, showReplies, showReactions, showZaps bool) string {
	var parts []string

	if showReplies && agg.ReplyCount > 0 {
		parts = append(parts, fmt.Sprintf("%d replies", agg.ReplyCount))
	}

	if showReactions && agg.ReactionTotal > 0 {
		// Show total reactions with breakdown
		if len(agg.ReactionCounts) > 0 {
			var reactionParts []string
			for emoji, count := range agg.ReactionCounts {
				reactionParts = append(reactionParts, fmt.Sprintf("%s %d", emoji, count))
			}
			parts = append(parts, fmt.Sprintf("%d reactions (%s)", agg.ReactionTotal, strings.Join(reactionParts, ", ")))
		} else {
			parts = append(parts, fmt.Sprintf("%d reactions", agg.ReactionTotal))
		}
	}

	if showZaps && agg.ZapSatsTotal > 0 {
		parts = append(parts, fmt.Sprintf("%s zapped", aggregates.FormatSats(agg.ZapSatsTotal)))
	}

	if len(parts) == 0 {
		return ""
	}

	return "Interactions: " + strings.Join(parts, ", ") + "\n"
}

// truncatePubkey truncates a pubkey for display
func truncatePubkey(pubkey string) string {
	if len(pubkey) <= 16 {
		return pubkey
	}
	return pubkey[:8] + "..." + pubkey[len(pubkey)-8:]
}

// formatTimestamp formats a Nostr timestamp
func formatTimestamp(ts nostr.Timestamp) string {
	t := time.Unix(int64(ts), 0)
	now := time.Now()

	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("2006-01-02 15:04")
}

// applyHeadersFooters wraps content with configured headers and footers
func (r *Renderer) applyHeadersFooters(content, page string) string {
	var sb strings.Builder

	// Add header if configured
	if header, err := r.loader.GetHeader(page); err == nil && header != "" {
		sb.WriteString(header)
		sb.WriteString("\n\n")
	}

	// Add main content
	sb.WriteString(content)

	// Add footer if configured
	if footer, err := r.loader.GetFooter(page); err == nil && footer != "" {
		sb.WriteString("\n\n")
		sb.WriteString(footer)
	}

	return sb.String()
}

// GetSummary creates a summary of content for display
func (r *Renderer) GetSummary(content string, maxLen int) string {
	// Remove newlines
	summary := strings.ReplaceAll(content, "\n", " ")
	summary = strings.ReplaceAll(summary, "\r", "")

	// Trim whitespace
	summary = strings.TrimSpace(summary)

	// Truncate if needed
	if len(summary) > maxLen {
		return summary[:maxLen] + "..."
	}

	return summary
}
