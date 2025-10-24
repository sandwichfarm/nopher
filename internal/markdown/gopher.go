package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// GopherRenderer renders markdown as plain text for Gopher protocol
type GopherRenderer struct {
	opts      *RenderOptions
	buf       *bytes.Buffer
	listDepth int
	linkRefs  []string
}

// NewGopherRenderer creates a new Gopher renderer
func NewGopherRenderer(opts *RenderOptions) *GopherRenderer {
	return &GopherRenderer{
		opts:     opts,
		buf:      &bytes.Buffer{},
		linkRefs: make([]string, 0),
	}
}

// Render renders the AST as plain text
func (r *GopherRenderer) Render(node ast.Node, source []byte) string {
	r.buf.Reset()
	r.linkRefs = r.linkRefs[:0]
	r.listDepth = 0

	r.renderNode(node, source)

	// Add link references at the end if using reference style
	if r.opts.LinkStyle == "reference" && len(r.linkRefs) > 0 {
		r.buf.WriteString("\n\nLinks:\n")
		for i, link := range r.linkRefs {
			r.buf.WriteString(fmt.Sprintf("[%d] %s\n", i+1, link))
		}
	}

	return r.buf.String()
}

func (r *GopherRenderer) renderNode(node ast.Node, source []byte) {
	WalkAST(node, source, func(n ast.Node, entering bool) ast.WalkStatus {
		return r.renderNodeInternal(n, source, entering)
	})
}

func (r *GopherRenderer) renderNodeInternal(n ast.Node, source []byte, entering bool) ast.WalkStatus {
	switch node := n.(type) {
	case *ast.Document:
		return ast.WalkContinue

	case *ast.Heading:
		if entering {
			r.buf.WriteString("\n")
			// Add header decoration
			level := node.Level
			if level == 1 {
				r.buf.WriteString("=== ")
			} else if level == 2 {
				r.buf.WriteString("--- ")
			} else {
				r.buf.WriteString(strings.Repeat("#", level) + " ")
			}
		} else {
			if node.Level == 1 || node.Level == 2 {
				text := ExtractText(node, source)
				r.buf.WriteString(" ")
				if node.Level == 1 {
					r.buf.WriteString(strings.Repeat("=", len(text)+8))
				} else {
					r.buf.WriteString(strings.Repeat("-", len(text)+8))
				}
			}
			r.buf.WriteString("\n\n")
		}
		return ast.WalkContinue

	case *ast.Paragraph:
		if entering {
			r.buf.WriteString(strings.Repeat(" ", r.listDepth*r.opts.IndentSize))
		} else {
			r.buf.WriteString("\n\n")
		}
		return ast.WalkContinue

	case *ast.Text:
		if entering {
			r.buf.Write(node.Text(source))
			if node.SoftLineBreak() {
				r.buf.WriteString(" ")
			} else if node.HardLineBreak() {
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
			// Handle link text
			return ast.WalkContinue
		} else {
			// Handle link URL after text
			if r.opts.PreserveLinks {
				linkURL := string(node.Destination)
				_ = ExtractText(node, source) // linkText already rendered in entering phase

				switch r.opts.LinkStyle {
				case "inline":
					r.buf.WriteString(fmt.Sprintf(" [%s]", linkURL))
				case "reference":
					r.linkRefs = append(r.linkRefs, linkURL)
					r.buf.WriteString(fmt.Sprintf("[%d]", len(r.linkRefs)))
				case "full":
					r.buf.WriteString(fmt.Sprintf(" (%s)", linkURL))
				default:
					// Just show the link text
				}
			}
		}
		return ast.WalkContinue

	case *ast.List:
		if entering {
			r.listDepth++
			if node.Start != 1 {
				r.buf.WriteString("\n")
			}
		} else {
			r.listDepth--
			r.buf.WriteString("\n")
		}
		return ast.WalkContinue

	case *ast.ListItem:
		if entering {
			indent := strings.Repeat(" ", (r.listDepth-1)*r.opts.IndentSize)
			r.buf.WriteString(indent)

			// Get parent list to determine marker
			parent := node.Parent()
			if list, ok := parent.(*ast.List); ok {
				if list.IsOrdered() {
					// TODO: Track item number
					r.buf.WriteString(fmt.Sprintf("%d. ", 1))
				} else {
					r.buf.WriteString("• ")
				}
			}
		}
		return ast.WalkContinue

	case *ast.CodeBlock, *ast.FencedCodeBlock:
		if entering {
			r.buf.WriteString("\n")
			lines := bytes.Split(node.Text(source), []byte("\n"))
			for _, line := range lines {
				r.buf.WriteString("    ")
				r.buf.Write(line)
				r.buf.WriteString("\n")
			}
			r.buf.WriteString("\n")
		}
		return ast.WalkSkipChildren

	case *ast.CodeSpan:
		if entering {
			r.buf.WriteString("`")
			r.buf.Write(node.Text(source))
			r.buf.WriteString("`")
		}
		return ast.WalkSkipChildren

	// Note: Emphasis and Strong nodes aren't in the base AST package
	// They're handled by child text nodes, formatting preserved in text

	case *ast.Blockquote:
		if entering {
			r.buf.WriteString("\n")
			lines := bytes.Split(ExtractTextBytes(node, source), []byte("\n"))
			for _, line := range lines {
				r.buf.WriteString("> ")
				r.buf.Write(line)
				r.buf.WriteString("\n")
			}
			r.buf.WriteString("\n")
		}
		return ast.WalkSkipChildren

	case *ast.ThematicBreak:
		if entering {
			r.buf.WriteString("\n")
			r.buf.WriteString(strings.Repeat("-", 70))
			r.buf.WriteString("\n\n")
		}
		return ast.WalkContinue

	default:
		// Unknown node type - continue walking
		return ast.WalkContinue
	}
}

// ExtractTextBytes extracts text as bytes
func ExtractTextBytes(node ast.Node, source []byte) []byte {
	return []byte(ExtractText(node, source))
}
