package gopher

import (
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nopher/internal/aggregates"
	"github.com/sandwich/nopher/internal/markdown"
)

// Renderer renders Nostr events as Gopher text
type Renderer struct {
	parser *markdown.Parser
}

// NewRenderer creates a new event renderer
func NewRenderer() *Renderer {
	return &Renderer{
		parser: markdown.NewParser(),
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

	// Content (render markdown)
	rendered, _ := r.parser.RenderGopher([]byte(event.Content), nil)
	sb.WriteString(rendered)

	// Aggregates footer
	if agg != nil && agg.HasInteractions() {
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("-", 70))
		sb.WriteString("\n")
		sb.WriteString(r.renderAggregates(agg))
	}

	return sb.String()
}

// RenderProfile renders a profile event
func (r *Renderer) RenderProfile(profileEvent *nostr.Event) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Profile: %s\n", truncatePubkey(profileEvent.PubKey)))
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n\n")

	// TODO: Parse kind 0 JSON content for name, about, picture
	sb.WriteString("Content:\n")
	sb.WriteString(profileEvent.Content)
	sb.WriteString("\n")

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

	for i, note := range notes {
		// Extract first line of content as summary
		content := note.Event.Content
		if len(content) > 70 {
			content = content[:67] + "..."
		}
		firstLine := strings.Split(content, "\n")[0]

		lines = append(lines, fmt.Sprintf("%d. %s", i+1, firstLine))
		lines = append(lines, fmt.Sprintf("   by %s - %s",
			truncatePubkey(note.Event.PubKey),
			formatTimestamp(note.Event.CreatedAt)))

		if note.Aggregates != nil && note.Aggregates.HasInteractions() {
			lines = append(lines, fmt.Sprintf("   %s", r.renderAggregates(note.Aggregates)))
		}

		lines = append(lines, "")
	}

	return lines
}
