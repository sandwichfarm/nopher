package gopher

import (
	"context"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nopher/internal/aggregates"
)

// Router handles selector routing for Gopher requests
type Router struct {
	server   *Server
	host     string
	port     int
	renderer *Renderer
}

// NewRouter creates a new router
func NewRouter(server *Server, host string, port int) *Router {
	return &Router{
		server:   server,
		host:     host,
		port:     port,
		renderer: NewRenderer(),
	}
}

// Route routes a selector to the appropriate handler
func (r *Router) Route(selector string) []byte {
	ctx := context.Background()

	// Empty selector = root/home
	if selector == "" || selector == "/" {
		return r.handleRoot(ctx)
	}

	// Parse selector path
	parts := strings.Split(strings.TrimPrefix(selector, "/"), "/")
	if len(parts) == 0 {
		return r.handleRoot(ctx)
	}

	section := parts[0]

	switch section {
	case "notes":
		return r.handleNotes(ctx, parts[1:])

	case "articles":
		return r.handleArticles(ctx, parts[1:])

	case "replies":
		return r.handleReplies(ctx, parts[1:])

	case "mentions":
		return r.handleMentions(ctx, parts[1:])

	case "note":
		if len(parts) >= 2 {
			return r.handleNote(ctx, parts[1])
		}
		return r.errorResponse("Missing note ID")

	case "thread":
		if len(parts) >= 2 {
			return r.handleThread(ctx, parts[1])
		}
		return r.errorResponse("Missing thread ID")

	case "profile":
		if len(parts) >= 2 {
			return r.handleProfile(ctx, parts[1])
		}
		return r.errorResponse("Missing pubkey")

	case "diagnostics":
		return r.handleDiagnostics(ctx)

	// Legacy support - redirect to new endpoints
	case "outbox":
		return r.handleNotes(ctx, parts[1:])

	case "inbox":
		return r.handleReplies(ctx, parts[1:])

	default:
		return r.errorResponse(fmt.Sprintf("Unknown selector: %s", selector))
	}
}

// handleRoot handles the root/home page
func (r *Router) handleRoot(ctx context.Context) []byte {
	gmap := NewGophermap(r.host, r.port)

	gmap.AddWelcome("Nopher - Nostr Gateway", "Browse Nostr content via Gopher protocol")

	gmap.AddDirectory("Notes", "/notes")
	gmap.AddDirectory("Articles", "/articles")
	gmap.AddDirectory("Replies", "/replies")
	gmap.AddDirectory("Mentions", "/mentions")
	gmap.AddSpacer()
	gmap.AddDirectory("Diagnostics", "/diagnostics")
	gmap.AddSpacer()
	gmap.AddInfo("Powered by Nopher")

	return gmap.Bytes()
}

// handleOutbox handles outbox listing
func (r *Router) handleOutbox(ctx context.Context, parts []string) []byte {
	gmap := NewGophermap(r.host, r.port)

	// Check if viewing a specific note
	if len(parts) > 0 && parts[0] != "" {
		return r.handleNote(ctx, parts[0])
	}

	// Query outbox notes
	queryHelper := r.server.GetQueryHelper()
	notes, err := queryHelper.GetOutboxNotes(ctx, 50)
	if err != nil {
		gmap.AddError(fmt.Sprintf("Error loading outbox: %v", err))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render note list
	lines := r.renderer.RenderNoteList(notes, "Outbox - My Notes")
	gmap.AddInfoBlock(lines)

	// Add note links
	if len(notes) > 0 {
		gmap.AddSpacer()
		gmap.AddInfo("Read notes:")
		gmap.AddSpacer()
		for i, note := range notes {
			// Extract first line for display
			content := note.Event.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			firstLine := strings.Split(content, "\n")[0]

			gmap.AddTextFile(
				fmt.Sprintf("%d. %s", i+1, firstLine),
				fmt.Sprintf("/outbox/%s", note.Event.ID),
			)
		}
	}

	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// handleInbox handles inbox listing (legacy - redirects to replies)
func (r *Router) handleInbox(ctx context.Context, parts []string) []byte {
	return r.handleReplies(ctx, parts)
}

// handleNotes handles notes listing (kind 1, non-replies)
func (r *Router) handleNotes(ctx context.Context, parts []string) []byte {
	gmap := NewGophermap(r.host, r.port)

	// Check if viewing a specific note
	if len(parts) > 0 && parts[0] != "" {
		return r.handleNote(ctx, parts[0])
	}

	// Query notes
	queryHelper := r.server.GetQueryHelper()
	notes, err := queryHelper.GetNotes(ctx, 50)
	if err != nil {
		gmap.AddError(fmt.Sprintf("Error loading notes: %v", err))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render note list with info
	lines := r.renderer.RenderNoteList(notes, "Notes")
	gmap.AddInfoBlock(lines)

	// Add clickable note links
	if len(notes) > 0 {
		gmap.AddSpacer()
		for i, note := range notes {
			// Extract first line for display
			content := note.Event.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			firstLine := strings.Split(content, "\n")[0]

			gmap.AddTextFile(
				fmt.Sprintf("%d. %s", i+1, firstLine),
				fmt.Sprintf("/note/%s", note.Event.ID),
			)
		}
	}

	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// handleArticles handles articles listing (kind 30023)
func (r *Router) handleArticles(ctx context.Context, parts []string) []byte {
	gmap := NewGophermap(r.host, r.port)

	// Query articles
	queryHelper := r.server.GetQueryHelper()
	articles, err := queryHelper.GetArticles(ctx, 50)
	if err != nil {
		gmap.AddError(fmt.Sprintf("Error loading articles: %v", err))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render article list
	lines := r.renderer.RenderNoteList(articles, "Articles")
	gmap.AddInfoBlock(lines)

	// Add article links
	if len(articles) > 0 {
		gmap.AddSpacer()
		for i, article := range articles {
			// Extract title or first line for display
			content := article.Event.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			firstLine := strings.Split(content, "\n")[0]

			gmap.AddTextFile(
				fmt.Sprintf("%d. %s", i+1, firstLine),
				fmt.Sprintf("/note/%s", article.Event.ID),
			)
		}
	}

	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// handleReplies handles replies listing
func (r *Router) handleReplies(ctx context.Context, parts []string) []byte {
	gmap := NewGophermap(r.host, r.port)

	// Query replies
	queryHelper := r.server.GetQueryHelper()
	replies, err := queryHelper.GetReplies(ctx, 50)
	if err != nil {
		gmap.AddError(fmt.Sprintf("Error loading replies: %v", err))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render reply list
	lines := r.renderer.RenderNoteList(replies, "Replies")
	gmap.AddInfoBlock(lines)

	// Add reply links
	if len(replies) > 0 {
		gmap.AddSpacer()
		for i, reply := range replies {
			// Extract first line for display
			content := reply.Event.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			firstLine := strings.Split(content, "\n")[0]

			gmap.AddTextFile(
				fmt.Sprintf("%d. %s", i+1, firstLine),
				fmt.Sprintf("/note/%s", reply.Event.ID),
			)
		}
	}

	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// handleMentions handles mentions listing
func (r *Router) handleMentions(ctx context.Context, parts []string) []byte {
	gmap := NewGophermap(r.host, r.port)

	// Query mentions
	queryHelper := r.server.GetQueryHelper()
	mentions, err := queryHelper.GetMentions(ctx, 50)
	if err != nil {
		gmap.AddError(fmt.Sprintf("Error loading mentions: %v", err))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render mention list
	lines := r.renderer.RenderNoteList(mentions, "Mentions")
	gmap.AddInfoBlock(lines)

	// Add mention links
	if len(mentions) > 0 {
		gmap.AddSpacer()
		for i, mention := range mentions {
			// Extract first line for display
			content := mention.Event.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			firstLine := strings.Split(content, "\n")[0]

			gmap.AddTextFile(
				fmt.Sprintf("%d. %s", i+1, firstLine),
				fmt.Sprintf("/note/%s", mention.Event.ID),
			)
		}
	}

	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// handleNote handles displaying a single note
func (r *Router) handleNote(ctx context.Context, noteID string) []byte {
	// Query the note
	events, err := r.server.GetStorage().QueryEvents(ctx, nostr.Filter{
		IDs: []string{noteID},
	})
	if err != nil || len(events) == 0 {
		gmap := NewGophermap(r.host, r.port)
		gmap.AddError(fmt.Sprintf("Note not found: %s", noteID))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	note := events[0]

	// Get aggregates from storage
	aggData, err := r.server.GetStorage().GetAggregate(ctx, noteID)
	var agg *aggregates.EventAggregates
	if err == nil && aggData != nil {
		agg = &aggregates.EventAggregates{
			EventID:         aggData.EventID,
			ReplyCount:      aggData.ReplyCount,
			ReactionTotal:   aggData.ReactionTotal,
			ReactionCounts:  aggData.ReactionCounts,
			ZapSatsTotal:    aggData.ZapSatsTotal,
			LastInteraction: aggData.LastInteractionAt,
		}
	}

	// Render the note as plain text
	text := r.renderer.RenderNote(note, agg)

	// Return as plain text with gopher terminator (not gophermap)
	return append([]byte(text), []byte(".\r\n")...)
}

// handleThread handles displaying a thread
func (r *Router) handleThread(ctx context.Context, rootID string) []byte {
	queryHelper := r.server.GetQueryHelper()

	// Query the thread
	thread, err := queryHelper.GetThreadByEvent(ctx, rootID)
	if err != nil || thread == nil {
		gmap := NewGophermap(r.host, r.port)
		gmap.AddError(fmt.Sprintf("Thread not found: %s", rootID))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	// Render the thread
	text := r.renderer.RenderThread(thread.Root, thread.Replies)

	// Return as plain text with gopher terminator
	return append([]byte(text), []byte(".\r\n")...)
}

// handleProfile handles displaying a profile
func (r *Router) handleProfile(ctx context.Context, pubkey string) []byte {
	// Query profile metadata (kind 0)
	events, err := r.server.GetStorage().QueryEvents(ctx, nostr.Filter{
		Kinds:   []int{0},
		Authors: []string{pubkey},
		Limit:   1,
	})
	if err != nil || len(events) == 0 {
		gmap := NewGophermap(r.host, r.port)
		gmap.AddError(fmt.Sprintf("Profile not found: %s", pubkey))
		gmap.AddSpacer()
		gmap.AddDirectory("← Back to Home", "/")
		return gmap.Bytes()
	}

	profile := events[0]

	// Render the profile
	text := r.renderer.RenderProfile(profile)

	// Return as plain text with gopher terminator
	return append([]byte(text), []byte(".\r\n")...)
}

// handleDiagnostics handles the diagnostics page
func (r *Router) handleDiagnostics(ctx context.Context) []byte {
	gmap := NewGophermap(r.host, r.port)

	gmap.AddInfo("Diagnostics")
	gmap.AddInfo(strings.Repeat("=", 15))
	gmap.AddSpacer()

	gmap.AddInfo("Server Status: Running")
	gmap.AddInfo(fmt.Sprintf("Host: %s", r.host))
	gmap.AddInfo(fmt.Sprintf("Port: %d", r.port))
	gmap.AddSpacer()

	// TODO: Add storage stats, sync status, etc.
	gmap.AddInfo("Storage: Connected")
	gmap.AddSpacer()

	gmap.AddDirectory("← Back to Home", "/")

	return gmap.Bytes()
}

// errorResponse returns an error gophermap
func (r *Router) errorResponse(message string) []byte {
	gmap := NewGophermap(r.host, r.port)
	gmap.AddError(message)
	gmap.AddSpacer()
	gmap.AddDirectory("← Back to Home", "/")
	return gmap.Bytes()
}
