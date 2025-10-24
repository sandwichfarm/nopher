package gemini

import (
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/markdown"
)

// Renderer renders Nostr events as Gemtext
type Renderer struct {
	parser *markdown.Parser
}

// NewRenderer creates a new event renderer
func NewRenderer() *Renderer {
	return &Renderer{
		parser: markdown.NewParser(),
	}
}

// RenderHome renders the home page
func (r *Renderer) RenderHome() string {
	var sb strings.Builder

	sb.WriteString("# Nopher - Nostr Gateway\n\n")
	sb.WriteString("Browse Nostr content via Gemini protocol\n\n")
	sb.WriteString("## Navigation\n\n")
	sb.WriteString("=> /outbox Outbox (My Notes)\n")
	sb.WriteString("=> /inbox Inbox (Replies & Mentions)\n")
	sb.WriteString("=> /search Search\n")
	sb.WriteString("=> /diagnostics Diagnostics\n")
	sb.WriteString("\n")
	sb.WriteString("Powered by Nopher\n")

	return sb.String()
}

// RenderNote renders a note event as gemtext
func (r *Renderer) RenderNote(event *nostr.Event, agg *aggregates.EventAggregates, threadURL, homeURL string) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Note by %s\n", truncatePubkey(event.PubKey)))
	sb.WriteString(fmt.Sprintf("Posted: %s\n\n", formatTimestamp(event.CreatedAt)))

	// Content (render markdown as gemtext)
	rendered, _ := r.parser.RenderGemini([]byte(event.Content), nil)
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

	sb.WriteString(fmt.Sprintf("# Profile: %s\n\n", truncatePubkey(profileEvent.PubKey)))

	// TODO: Parse kind 0 JSON content for name, about, picture
	sb.WriteString("## Content\n\n")
	sb.WriteString("```\n")
	sb.WriteString(profileEvent.Content)
	sb.WriteString("\n```\n\n")

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

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(notes) == 0 {
		sb.WriteString("No notes yet.\n\n")
		sb.WriteString(fmt.Sprintf("=> %s Back to Home\n", homeURL))
		return sb.String()
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

	return sb.String()
}

// renderAggregates renders interaction stats
func (r *Renderer) renderAggregates(agg *aggregates.EventAggregates) string {
	var parts []string

	if agg.ReplyCount > 0 {
		parts = append(parts, fmt.Sprintf("%d replies", agg.ReplyCount))
	}

	if agg.ReactionTotal > 0 {
		parts = append(parts, fmt.Sprintf("%d reactions", agg.ReactionTotal))
	}

	if agg.ZapSatsTotal > 0 {
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
