package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// FingerRenderer renders markdown as ultra-compact text for Finger protocol
type FingerRenderer struct {
	opts *RenderOptions
	buf  *bytes.Buffer
}

// NewFingerRenderer creates a new Finger renderer
func NewFingerRenderer(opts *RenderOptions) *FingerRenderer {
	return &FingerRenderer{
		opts: opts,
		buf:  &bytes.Buffer{},
	}
}

// Render renders the AST as compact text
func (r *FingerRenderer) Render(node ast.Node, source []byte) string {
	r.buf.Reset()

	// Extract all text, stripping formatting
	text := ExtractText(node, source)

	if r.opts.CompactMode {
		// Collapse whitespace
		text = strings.Join(strings.Fields(text), " ")
	}

	// Apply width limit if specified
	if r.opts.Width > 0 && len(text) > r.opts.Width {
		text = text[:r.opts.Width-3] + "..."
	}

	r.buf.WriteString(text)

	return r.buf.String()
}
