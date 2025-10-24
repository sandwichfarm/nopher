package gopher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/config"
	"github.com/sandwich/nopher/internal/entities"
	"github.com/sandwich/nopher/internal/markdown"
	nostrclient "github.com/sandwich/nopher/internal/nostr"
	"github.com/sandwich/nopher/internal/presentation"
	"github.com/sandwich/nopher/internal/storage"
)

// Renderer renders Nostr events as Gopher text
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

// RenderNote renders a note event as plain text
func (r *Renderer) RenderNote(event *nostr.Event, agg *aggregates.EventAggregates) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Note by %s\n", truncatePubkey(event.PubKey)))
	sb.WriteString(fmt.Sprintf("Posted: %s\n", formatTimestamp(event.CreatedAt)))
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n\n")

	// Content (resolve NIP-19 entities, then render markdown)
	content := event.Content

	// Resolve NIP-19 entities
	ctx := context.Background()
	content = r.resolver.ReplaceEntities(ctx, content, entities.GopherFormatter)

	// Apply max content length if configured
	if r.config.Display.Limits.MaxContentLength > 0 && len(content) > r.config.Display.Limits.MaxContentLength {
		content = content[:r.config.Display.Limits.MaxContentLength] + r.config.Display.Limits.TruncateIndicator
	}

	rendered, _ := r.parser.RenderGopher([]byte(content), nil)
	sb.WriteString(rendered)

	// Aggregates footer - only show if configured for detail view
	if r.config.Display.Detail.ShowInteractions && agg != nil && agg.HasInteractions() {
		sb.WriteString("\n")
		sb.WriteString(r.applyConfigSeparator("section"))
		sb.WriteString("\n")
		sb.WriteString(r.renderAggregatesForDetail(agg))
	}

	return sb.String()
}

// RenderProfile renders a profile event
func (r *Renderer) RenderProfile(profileEvent *nostr.Event) string {
	var sb strings.Builder

	// Parse profile metadata
	profile := nostrclient.ParseProfile(profileEvent)
	if profile == nil {
		// Fallback for invalid profile
		sb.WriteString(fmt.Sprintf("Profile: %s\n", truncatePubkey(profileEvent.PubKey)))
		sb.WriteString(strings.Repeat("=", 70))
		sb.WriteString("\n\nInvalid profile data\n")
		return sb.String()
	}

	// Header with display name
	displayName := profile.GetDisplayName()
	if displayName == "" {
		displayName = truncatePubkey(profileEvent.PubKey)
	}

	sb.WriteString(fmt.Sprintf("Profile: %s\n", displayName))
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n\n")

	// Pubkey
	sb.WriteString(fmt.Sprintf("Pubkey: %s\n", profileEvent.PubKey))
	sb.WriteString("\n")

	// Name fields
	if profile.Name != "" {
		sb.WriteString(fmt.Sprintf("Name: %s\n", profile.Name))
	}
	if profile.DisplayName != "" && profile.DisplayName != profile.Name {
		sb.WriteString(fmt.Sprintf("Display Name: %s\n", profile.DisplayName))
	}

	// About/Bio
	if profile.About != "" {
		sb.WriteString("\nAbout:\n")
		sb.WriteString(profile.About)
		sb.WriteString("\n")
	}

	// Contact information
	if profile.Website != "" {
		sb.WriteString(fmt.Sprintf("\nWebsite: %s\n", profile.Website))
	}
	if profile.NIP05 != "" {
		sb.WriteString(fmt.Sprintf("NIP-05: %s\n", profile.NIP05))
	}

	// Lightning info
	lightningAddr := profile.GetLightningAddress()
	if lightningAddr != "" {
		sb.WriteString(fmt.Sprintf("Lightning: %s\n", lightningAddr))
	}

	// Media
	if profile.Picture != "" {
		sb.WriteString(fmt.Sprintf("\nPicture: %s\n", profile.Picture))
	}
	if profile.Banner != "" {
		sb.WriteString(fmt.Sprintf("Banner: %s\n", profile.Banner))
	}

	return sb.String()
}

// RenderThread renders a thread with indentation
func (r *Renderer) RenderThread(root *aggregates.EnrichedEvent, replies []*aggregates.EnrichedEvent) string {
	var sb strings.Builder

	sb.WriteString("Thread\n")
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n\n")

	// Root post
	sb.WriteString("● Root Post\n")
	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n")
	sb.WriteString(r.RenderNote(root.Event, root.Aggregates))
	sb.WriteString("\n\n")

	// Replies
	if len(replies) > 0 {
		sb.WriteString(fmt.Sprintf("Replies (%d)\n", len(replies)))
		sb.WriteString(strings.Repeat("-", 70))
		sb.WriteString("\n\n")

		for i, reply := range replies {
			sb.WriteString(fmt.Sprintf("  ↳ Reply %d by %s\n", i+1, truncatePubkey(reply.Event.PubKey)))
			sb.WriteString(fmt.Sprintf("    %s\n\n", formatTimestamp(reply.Event.CreatedAt)))

			// Indent reply content
			content, _ := r.parser.RenderGopher([]byte(reply.Event.Content), nil)
			indented := indentText(content, "    ")
			sb.WriteString(indented)
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("No replies yet.\n")
	}

	return sb.String()
}

// renderAggregates renders interaction stats (for feed view - respects feed config)
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

// applyConfigSeparator applies the configured separator for the given type
func (r *Renderer) applyConfigSeparator(separatorType string) string {
	var sep string
	switch separatorType {
	case "item":
		sep = r.config.Presentation.Separators.Item.Gopher
	case "section":
		sep = r.config.Presentation.Separators.Section.Gopher
	default:
		sep = "---"
	}

	// If no custom separator, use default visual separator
	if sep == "" {
		if separatorType == "section" {
			return strings.Repeat("-", 70)
		}
		return ""
	}

	return sep
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

// indentText indents each line of text
func indentText(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// RenderNoteList renders a list of notes with summaries
func (r *Renderer) RenderNoteList(notes []*aggregates.EnrichedEvent, title string) []string {
	lines := make([]string, 0)

	lines = append(lines, title)
	lines = append(lines, strings.Repeat("=", len(title)))
	lines = append(lines, "")

	if len(notes) == 0 {
		lines = append(lines, "No notes yet")
		return lines
	}

	summaryLength := r.config.Display.Limits.SummaryLength
	if summaryLength <= 0 {
		summaryLength = 70 // Default fallback
	}

	for i, note := range notes {
		// Extract first line of content as summary
		content := note.Event.Content
		if len(content) > summaryLength {
			content = content[:summaryLength-len(r.config.Display.Limits.TruncateIndicator)] + r.config.Display.Limits.TruncateIndicator
		}
		firstLine := strings.Split(content, "\n")[0]

		lines = append(lines, fmt.Sprintf("%d. %s", i+1, firstLine))
		lines = append(lines, fmt.Sprintf("   by %s - %s",
			truncatePubkey(note.Event.PubKey),
			formatTimestamp(note.Event.CreatedAt)))

		// Only show aggregates if configured for feed view
		if r.config.Display.Feed.ShowInteractions && note.Aggregates != nil && note.Aggregates.HasInteractions() {
			aggStr := r.renderAggregates(note.Aggregates)
			if aggStr != "" {
				lines = append(lines, fmt.Sprintf("   %s", aggStr))
			}
		}

		// Apply item separator if configured
		itemSep := r.applyConfigSeparator("item")
		if itemSep != "" {
			lines = append(lines, itemSep)
		}

		lines = append(lines, "")
	}

	return lines
}
