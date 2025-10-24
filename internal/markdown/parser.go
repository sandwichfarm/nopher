package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser wraps goldmark for markdown parsing
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new markdown parser
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown (tables, strikethrough, etc.)
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
	)

	return &Parser{
		md: md,
	}
}

// Parse parses markdown into an AST
func (p *Parser) Parse(source []byte) ast.Node {
	reader := text.NewReader(source)
	doc := p.md.Parser().Parse(reader)
	return doc
}

// RenderGopher renders the AST as plain text for Gopher
func (p *Parser) RenderGopher(source []byte, opts *RenderOptions) (string, error) {
	if opts == nil {
		opts = DefaultGopherOptions()
	}

	doc := p.Parse(source)
	renderer := NewGopherRenderer(opts)
	return renderer.Render(doc, source), nil
}

// RenderGemini renders the AST as gemtext for Gemini
func (p *Parser) RenderGemini(source []byte, opts *RenderOptions) (string, error) {
	if opts == nil {
		opts = DefaultGeminiOptions()
	}

	doc := p.Parse(source)
	renderer := NewGeminiRenderer(opts)
	return renderer.Render(doc, source), nil
}

// RenderFinger renders the AST as compact text for Finger
func (p *Parser) RenderFinger(source []byte, opts *RenderOptions) (string, error) {
	if opts == nil {
		opts = DefaultFingerOptions()
	}

	doc := p.Parse(source)
	renderer := NewFingerRenderer(opts)
	return renderer.Render(doc, source), nil
}

// RenderOptions contains configuration for rendering
type RenderOptions struct {
	// Width is the maximum line width (0 = no wrapping)
	Width int

	// IndentSize is the number of spaces per indent level
	IndentSize int

	// PreserveLinks determines if links are preserved or stripped
	PreserveLinks bool

	// LinkStyle determines how links are rendered
	// "inline" - [text](url)
	// "reference" - [text][1] with footnotes
	// "full" - text (url)
	LinkStyle string

	// CompactMode removes extra whitespace
	CompactMode bool

	// StripFormatting removes all formatting (bold, italic, etc.)
	StripFormatting bool
}

// DefaultGopherOptions returns default options for Gopher rendering
func DefaultGopherOptions() *RenderOptions {
	return &RenderOptions{
		Width:           70,
		IndentSize:      2,
		PreserveLinks:   true,
		LinkStyle:       "reference",
		CompactMode:     false,
		StripFormatting: false,
	}
}

// DefaultGeminiOptions returns default options for Gemini rendering
func DefaultGeminiOptions() *RenderOptions {
	return &RenderOptions{
		Width:           0, // No wrapping for Gemini
		IndentSize:      0,
		PreserveLinks:   true,
		LinkStyle:       "gemini",
		CompactMode:     false,
		StripFormatting: false,
	}
}

// DefaultFingerOptions returns default options for Finger rendering
func DefaultFingerOptions() *RenderOptions {
	return &RenderOptions{
		Width:           80,
		IndentSize:      0,
		PreserveLinks:   false,
		LinkStyle:       "stripped",
		CompactMode:     true,
		StripFormatting: true,
	}
}

// WalkAST walks the AST and calls the visitor for each node
func WalkAST(node ast.Node, source []byte, visitor func(n ast.Node, entering bool) ast.WalkStatus) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		status := visitor(n, entering)
		return status, nil
	})
}

// ExtractText extracts plain text from an AST node
func ExtractText(node ast.Node, source []byte) string {
	var buf bytes.Buffer

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Text:
			buf.Write(n.Text(source))
		case *ast.String:
			buf.Write(n.Value)
		}

		return ast.WalkContinue, nil
	})

	return buf.String()
}
