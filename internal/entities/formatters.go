package entities

import "fmt"

// GopherFormatter formats an entity for Gopher protocol
// Returns inline text representation (Gopher doesn't support inline links)
func GopherFormatter(entity *Entity) string {
	// For Gopher, we can't embed clickable links inline
	// So we just show the display name with a reference marker
	return fmt.Sprintf("@%s", entity.DisplayName)
}

// GeminiFormatter formats an entity for Gemini protocol
// Returns a Gemini-style link
func GeminiFormatter(entity *Entity) string {
	// Gemini supports inline links like => URL text
	// But inline links within paragraphs aren't standard
	// So we use a similar approach to Gopher but with the link path shown
	return fmt.Sprintf("[%s](gemini://HOSTNAME%s)", entity.DisplayName, entity.Link)
}

// PlainTextFormatter formats an entity as plain text with display name
func PlainTextFormatter(entity *Entity) string {
	return entity.DisplayName
}

// MarkdownFormatter formats an entity as Markdown link
func MarkdownFormatter(entity *Entity) string {
	return fmt.Sprintf("[%s](%s)", entity.DisplayName, entity.Link)
}

// HTMLFormatter formats an entity as HTML link
func HTMLFormatter(entity *Entity) string {
	return fmt.Sprintf(`<a href="%s">%s</a>`, entity.Link, entity.DisplayName)
}
