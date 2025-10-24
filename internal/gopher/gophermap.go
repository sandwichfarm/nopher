package gopher

import (
	"fmt"
	"strings"
)

// ItemType represents a Gopher menu item type per RFC 1436
type ItemType byte

const (
	// RFC 1436 standard types
	ItemTypeTextFile   ItemType = '0' // Text file
	ItemTypeDirectory  ItemType = '1' // Directory/menu
	ItemTypeCSOServer  ItemType = '2' // CSO phone-book server
	ItemTypeError      ItemType = '3' // Error
	ItemTypeBinHex     ItemType = '4' // BinHexed Macintosh file
	ItemTypeDOSArchive ItemType = '5' // DOS binary archive
	ItemTypeUUEncoded  ItemType = '6' // UNIX uuencoded file
	ItemTypeSearch     ItemType = '7' // Index-Search server
	ItemTypeTelnet     ItemType = '8' // Telnet session
	ItemTypeBinary     ItemType = '9' // Binary file
	ItemTypeGIF        ItemType = 'g' // GIF format graphics file
	ItemTypeImage      ItemType = 'I' // Some kind of image file
	ItemTypeTelnet3270 ItemType = 'T' // Telnet3270 session

	// Non-standard but widely supported
	ItemTypeHTML ItemType = 'h' // HTML file
	ItemTypeInfo ItemType = 'i' // Informational message (non-selectable)
)

// Item represents a single line in a Gophermap
type Item struct {
	Type     ItemType
	Display  string // User display string
	Selector string // Selector string
	Host     string // Hostname
	Port     int    // Port number
}

// String formats an Item as a gophermap line per RFC 1436
// Format: Type + Display + TAB + Selector + TAB + Host + TAB + Port + CRLF
func (i *Item) String() string {
	return fmt.Sprintf("%c%s\t%s\t%s\t%d\r\n",
		i.Type,
		i.Display,
		i.Selector,
		i.Host,
		i.Port,
	)
}

// Gophermap represents a collection of menu items
type Gophermap struct {
	Items []Item
	host  string
	port  int
}

// NewGophermap creates a new gophermap with default host/port
func NewGophermap(host string, port int) *Gophermap {
	return &Gophermap{
		Items: make([]Item, 0),
		host:  host,
		port:  port,
	}
}

// AddItem adds an item to the gophermap
func (g *Gophermap) AddItem(itemType ItemType, display, selector string) {
	g.Items = append(g.Items, Item{
		Type:     itemType,
		Display:  display,
		Selector: selector,
		Host:     g.host,
		Port:     g.port,
	})
}

// AddInfo adds an informational (non-selectable) line
func (g *Gophermap) AddInfo(text string) {
	g.AddItem(ItemTypeInfo, text, "fake")
}

// AddDirectory adds a directory/submenu item
func (g *Gophermap) AddDirectory(display, selector string) {
	g.AddItem(ItemTypeDirectory, display, selector)
}

// AddTextFile adds a text file item
func (g *Gophermap) AddTextFile(display, selector string) {
	g.AddItem(ItemTypeTextFile, display, selector)
}

// AddError adds an error item
func (g *Gophermap) AddError(message string) {
	g.AddItem(ItemTypeError, message, "error")
}

// AddSpacer adds a blank line for visual separation
func (g *Gophermap) AddSpacer() {
	g.AddInfo("")
}

// String renders the complete gophermap
func (g *Gophermap) String() string {
	var sb strings.Builder

	for _, item := range g.Items {
		sb.WriteString(item.String())
	}

	// Add trailing period (end-of-transmission marker)
	sb.WriteString(".\r\n")

	return sb.String()
}

// Bytes returns the gophermap as bytes
func (g *Gophermap) Bytes() []byte {
	return []byte(g.String())
}

// AddInfoBlock adds multiple lines of informational text
func (g *Gophermap) AddInfoBlock(lines []string) {
	for _, line := range lines {
		g.AddInfo(line)
	}
}

// AddWelcome adds a standard welcome section
func (g *Gophermap) AddWelcome(title, description string) {
	g.AddSpacer()
	g.AddInfo(title)
	g.AddInfo(strings.Repeat("=", len(title)))
	g.AddSpacer()
	if description != "" {
		// Word wrap description
		wrapped := wrapText(description, 70)
		g.AddInfoBlock(wrapped)
		g.AddSpacer()
	}
}

// wrapText wraps text to the specified width
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= width {
			current.WriteString(" ")
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}
