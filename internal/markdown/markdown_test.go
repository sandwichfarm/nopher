package markdown

import (
	"strings"
	"testing"
)

const sampleMarkdown = `# Main Heading

This is a paragraph with **bold** and *italic* text.

## Subheading

Here's a [link](https://example.com) and some ` + "`inline code`" + `.

### List Example

* Item 1
* Item 2
* Item 3

Ordered list:

1. First
2. Second
3. Third

` + "```" + `
code block
with multiple lines
` + "```" + `

> This is a blockquote
> with multiple lines

---

Final paragraph.
`

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("Expected parser, got nil")
	}

	if p.md == nil {
		t.Error("Parser markdown instance not initialized")
	}
}

func TestParse(t *testing.T) {
	p := NewParser()
	source := []byte("# Hello World")

	node := p.Parse(source)
	if node == nil {
		t.Fatal("Expected AST node, got nil")
	}
}

func TestRenderGopher(t *testing.T) {
	p := NewParser()
	output, err := p.RenderGopher([]byte(sampleMarkdown), nil)
	if err != nil {
		t.Fatalf("RenderGopher() error = %v", err)
	}

	// Check for key elements
	if !strings.Contains(output, "Main Heading") {
		t.Error("Output missing main heading")
	}

	if !strings.Contains(output, "bold") {
		t.Error("Output missing bold text content")
	}

	if !strings.Contains(output, "• Item 1") {
		t.Error("Output missing list item")
	}

	if !strings.Contains(output, "Links:") {
		t.Error("Output missing link references section")
	}
}

func TestRenderGopherWithOptions(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name string
		opts *RenderOptions
		want string
	}{
		{
			name: "inline links",
			opts: &RenderOptions{
				Width:         70,
				PreserveLinks: true,
				LinkStyle:     "inline",
			},
			want: "https://example.com",
		},
		{
			name: "stripped links",
			opts: &RenderOptions{
				Width:         70,
				PreserveLinks: false,
			},
			want: "link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := p.RenderGopher([]byte(sampleMarkdown), tt.opts)
			if err != nil {
				t.Fatalf("RenderGopher() error = %v", err)
			}

			if tt.opts.PreserveLinks && !strings.Contains(output, tt.want) {
				t.Errorf("Expected output to contain %q", tt.want)
			}
		})
	}
}

func TestRenderGemini(t *testing.T) {
	p := NewParser()
	output, err := p.RenderGemini([]byte(sampleMarkdown), nil)
	if err != nil {
		t.Fatalf("RenderGemini() error = %v", err)
	}

	// Check for gemtext syntax
	if !strings.Contains(output, "# Main Heading") {
		t.Error("Output missing gemini heading")
	}

	if !strings.Contains(output, "* Item 1") {
		t.Error("Output missing gemini list item")
	}

	if !strings.Contains(output, "=> https://example.com") {
		t.Error("Output missing gemini link")
	}

	if !strings.Contains(output, "```") {
		t.Error("Output missing gemini code block markers")
	}
}

func TestRenderFinger(t *testing.T) {
	p := NewParser()
	output, err := p.RenderFinger([]byte(sampleMarkdown), nil)
	if err != nil {
		t.Fatalf("RenderFinger() error = %v", err)
	}

	// Finger output should be compact
	if strings.Count(output, "\n") > 2 {
		t.Error("Finger output not compact enough")
	}

	// Should contain text but no formatting
	if !strings.Contains(output, "Main Heading") {
		t.Error("Output missing heading text")
	}

	if strings.Contains(output, "**") || strings.Contains(output, "*") {
		t.Error("Finger output should strip formatting")
	}
}

func TestRenderFingerWithWidthLimit(t *testing.T) {
	p := NewParser()
	opts := &RenderOptions{
		Width:       50,
		CompactMode: true,
	}

	output, err := p.RenderFinger([]byte(sampleMarkdown), opts)
	if err != nil {
		t.Fatalf("RenderFinger() error = %v", err)
	}

	if len(output) > 50 {
		t.Errorf("Output exceeds width limit: %d > 50", len(output))
	}

	if !strings.HasSuffix(output, "...") {
		t.Error("Truncated output should end with ...")
	}
}

func TestExtractText(t *testing.T) {
	p := NewParser()
	source := []byte("# Heading\n\nParagraph with **bold**.")

	node := p.Parse(source)
	text := ExtractText(node, source)

	if !strings.Contains(text, "Heading") {
		t.Error("Extracted text missing heading")
	}

	if !strings.Contains(text, "Paragraph") {
		t.Error("Extracted text missing paragraph")
	}

	if !strings.Contains(text, "bold") {
		t.Error("Extracted text missing bold content")
	}
}

func TestDefaultOptions(t *testing.T) {
	gopherOpts := DefaultGopherOptions()
	if gopherOpts.Width != 70 {
		t.Errorf("Expected Gopher width 70, got %d", gopherOpts.Width)
	}

	geminiOpts := DefaultGeminiOptions()
	if geminiOpts.Width != 0 {
		t.Errorf("Expected Gemini width 0, got %d", geminiOpts.Width)
	}

	fingerOpts := DefaultFingerOptions()
	if !fingerOpts.CompactMode {
		t.Error("Expected Finger compact mode enabled")
	}

	if !fingerOpts.StripFormatting {
		t.Error("Expected Finger strip formatting enabled")
	}
}

func TestRenderCodeBlock(t *testing.T) {
	p := NewParser()
	source := []byte("```go\nfunc main() {\n  fmt.Println(\"Hello\")\n}\n```")

	gopherOut, _ := p.RenderGopher(source, nil)
	if !strings.Contains(gopherOut, "func main") {
		t.Error("Gopher output missing code block content")
	}

	geminiOut, _ := p.RenderGemini(source, nil)
	if !strings.Contains(geminiOut, "```") {
		t.Error("Gemini output missing code block markers")
	}
}

func TestRenderBlockquote(t *testing.T) {
	p := NewParser()
	source := []byte("> This is a quote\n> Second line")

	gopherOut, _ := p.RenderGopher(source, nil)
	if !strings.Contains(gopherOut, "> This is a quote") {
		t.Error("Gopher output missing blockquote")
	}

	geminiOut, _ := p.RenderGemini(source, nil)
	if !strings.Contains(geminiOut, "> This is a quote") {
		t.Error("Gemini output missing blockquote")
	}
}
