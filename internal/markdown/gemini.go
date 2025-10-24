package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// GeminiRenderer renders markdown as gemtext for Gemini protocol
type GeminiRenderer struct {
	opts       *RenderOptions
	buf        *bytes.Buffer
	inCodeBlock bool
	inList      bool
}

// NewGeminiRenderer creates a new Gemini renderer
func NewGeminiRenderer(opts *RenderOptions) *GeminiRenderer {
	return &GeminiRenderer{
		opts: opts,
		buf:  &bytes.Buffer{},
	}
}

// Render renders the AST as gemtext
func (r *GeminiRenderer) Render(node ast.Node, source []byte) string {
	r.buf.Reset()
	r.inCodeBlock = false
	r.inList = false

	r.renderNode(node, source)

	return r.buf.String()
}

func (r *GeminiRenderer) renderNode(node ast.Node, source []byte) {
	WalkAST(node, source, func(n ast.Node, entering bool) ast.WalkStatus {
		return r.renderNodeInternal(n, source, entering)
	})
}

func (r *GeminiRenderer) renderNodeInternal(n ast.Node, source []byte, entering bool) ast.WalkStatus {
	switch node := n.(type) {
	case *ast.Document:
		return ast.WalkContinue

	case *ast.Heading:
		if entering {
			// Gemini headings: # ## ###
			level := node.Level
			if level > 3 {
				level = 3 // Gemini only supports 3 levels
			}
			r.buf.WriteString(strings.Repeat("#", level) + " ")
		} else {
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	case *ast.Paragraph:
		if !entering {
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	case *ast.Text:
		if entering {
			r.buf.Write(node.Text(source))
			if node.HardLineBreak() {
				r.buf.WriteString("\n")
			}
		}
		return ast.WalkContinue

	case *ast.String:
		if entering {
			r.buf.Write(node.Value)
		}
		return ast.WalkContinue

	case *ast.Link:
		if entering {
			// Gemini links are on their own line
			linkText := ExtractText(node, source)

			// Don't render link inline - save for after paragraph
			// For now, just show the text inline
			r.buf.WriteString(linkText)
		} else {
			// After the link node, add a gemini link line
			linkURL := string(node.Destination)
			linkText := ExtractText(node, source)
			r.buf.WriteString(fmt.Sprintf("\n=> %s %s\n", linkURL, linkText))
		}
		return ast.WalkSkipChildren

	case *ast.List:
		if entering {
			r.inList = true
		} else {
			r.inList = false
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	case *ast.ListItem:
		if entering {
			r.buf.WriteString("* ")
		} else {
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	case *ast.CodeBlock, *ast.FencedCodeBlock:
		if entering {
			// Gemini preformatted text
			r.buf.WriteString("```\n")
			r.buf.Write(node.Text(source))
			if !bytes.HasSuffix(node.Text(source), []byte("\n")) {
				r.buf.WriteString("\n")
			}
			r.buf.WriteString("```\n")
		}
		return ast.WalkSkipChildren

	case *ast.CodeSpan:
		if entering {
			// Gemini doesn't have inline code, just output as-is
			r.buf.Write(node.Text(source))
		}
		return ast.WalkSkipChildren

	// Note: Emphasis and Strong are handled by continuing to child text nodes
	// Gemini doesn't support these anyway

	case *ast.Blockquote:
		if entering {
			// Gemini doesn't have blockquotes, render as indented text
			r.buf.WriteString("> ")
			text := ExtractText(node, source)
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if i > 0 {
					r.buf.WriteString("> ")
				}
				r.buf.WriteString(line)
				r.buf.WriteString("\n")
			}
		}
		return ast.WalkSkipChildren

	case *ast.ThematicBreak:
		if entering {
			// Gemini doesn't have horizontal rules, use blank line
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	default:
		// Unknown node type - continue walking
		return ast.WalkContinue
	}
}
