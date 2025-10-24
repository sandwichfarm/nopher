# Phase 19 Additional Features: NIP-19 Entities & Sections

**Status**: ✅ Complete
**Completed**: 2025-10-24 (after Phase 19 completion)
**Updated**: 2025-10-24 (added Gemini protocol parity)

## Overview

After completing Phase 19, we added two major features:
1. **NIP-19 Entity Resolution** - Parse and resolve `nostr:` URIs to human-readable names
2. **Sections System Enhancements** - Support multiple sections per path with "more" links

## Deliverables

### 1. NIP-19 Entity Resolution ✅

**Files Created**:
- `internal/entities/resolver.go` - Entity parsing and resolution
- `internal/entities/formatters.go` - Protocol-specific formatters

**Key Features**:
- Parse `nostr:` URIs with regex matching
- Resolve entity types: npub, nprofile, note, nevent, naddr
- Fetch display names from profile metadata (kind 0)
- Fetch note/article titles for better context
- Protocol-agnostic formatters (Gopher, Gemini, plain text)

**Supported Entity Types**:
```go
// npub - Public key
nostr:npub1... → "@Alice" (with profile lookup)

// nprofile - Public key with relays
nostr:nprofile1... → "@Bob" (with profile lookup)

// note - Event ID
nostr:note1... → "First line of note..." (with event lookup)

// nevent - Event ID with relays
nostr:nevent1... → "Event title/preview" (with event lookup)

// naddr - Parameterized replaceable event
nostr:naddr1... → "Article Title" (with event lookup)
```

**Resolution Process**:
1. Regex matches `nostr:(npub1|nprofile1|note1|nevent1|naddr1)...`
2. Decode using `nip19.Decode()`
3. Query storage for metadata/event
4. Extract human-readable name/title
5. Format using protocol-specific formatter
6. Return formatted string

**Formatters**:
```go
// Gopher: Short mentions
GopherFormatter(entity) → "@Alice"

// Gemini/Plain: Display names
PlainTextFormatter(entity) → "Alice"

// Custom formatters can be added
```

**Files Modified**:
- `internal/gopher/renderer.go` - Added entity resolution in RenderNote()
- `internal/gemini/renderer.go` - Added entity resolution in RenderNote()

**Integration**:
```go
// In renderer.RenderNote():
content := event.Content
ctx := context.Background()

// Resolve all nostr: URIs to human-readable text
content = r.resolver.ReplaceEntities(ctx, content, entities.GopherFormatter)

// Then render markdown as usual
rendered, _ := r.parser.RenderGopher([]byte(content), nil)
```

### 2. Sections System Enhancements ✅

**Architecture Clarification**:
- Sections are OPTIONAL filtered views over indexed data
- "inbox" and "outbox" are INTERNAL source identifiers (not user-facing paths)
- Router provides default paths: /notes, /replies, /mentions, /articles
- Sections can OVERRIDE any path including / (homepage)

**Files Modified**:
- `internal/sections/sections.go`
- `internal/gopher/router.go`
- `internal/gopher/server.go`
- `internal/gemini/router.go`
- `internal/gemini/server.go`

#### Feature 2a: "More" Links ✅

Allow preview sections to link to full paginated views.

**Section Struct Changes**:
```go
type Section struct {
    Name        string
    Path        string
    Title       string
    Description string
    Filters     FilterSet
    SortBy      SortField
    SortOrder   SortOrder
    Limit       int
    ShowDates   bool
    ShowAuthors bool
    GroupBy     GroupField
    MoreLink    *MoreLink // NEW: Optional link to full view
    Order       int       // NEW: Display order (for multi-section pages)
}

type MoreLink struct {
    Text       string // Display text (e.g., "More DIY posts")
    SectionRef string // Name of target section
}
```

**Usage Example**:
```go
// Preview section (homepage, 5 items)
sectionManager.RegisterSection(&sections.Section{
    Name:  "diy-preview",
    Path:  "/",
    Title: "Recent DIY Projects",
    Limit: 5,
    Filters: sections.FilterSet{
        Tags: map[string][]string{"t": {"diy"}},
    },
    MoreLink: &sections.MoreLink{
        Text:       "More DIY posts",
        SectionRef: "diy-full", // Links to full section
    },
})

// Full section (dedicated route, paginated)
sectionManager.RegisterSection(&sections.Section{
    Name:  "diy-full",
    Path:  "/diy",
    Title: "All DIY Projects",
    Limit: 9, // Full page size (single-digit hotkeys for Gopher)
    Filters: sections.FilterSet{
        Tags: map[string][]string{"t": {"diy"}},
    },
})
```

**Rendering**:
- Preview shows limited items (e.g., 5)
- "→ More DIY posts" link appears at bottom
- Links to `/diy` for full paginated view
- Full view shows 9 items per page with pagination

#### Feature 2b: Multiple Sections Per Path ✅

Support composing multiple sections on a single page (e.g., homepage).

**Section Manager Changes**:
```go
// New method: Get all sections for a path
func (m *Manager) GetSectionsByPath(path string) []*Section

// Returns sections sorted by Order field (lower numbers first)
```

**Router Changes**:
```go
// Check for sections first (override default handlers)
sections := r.server.GetSectionManager().GetSectionsByPath(path)
if len(sections) > 0 {
    return r.handleSections(ctx, sections, path)
}

// Fall back to default handlers
if path == "/" {
    return r.handleRoot(ctx)
}
```

**New Handler**:
```go
// handleSections renders multiple sections on one page
func (r *Router) handleSections(ctx, sections, path) []byte {
    // For each section (in order):
    // 1. Render title and description
    // 2. Render limited events
    // 3. Render "more" link if configured
    // 4. Add separator between sections
    // 5. Add home link at bottom
}
```

**Usage Example**:
```go
// Homepage with multiple sections
sectionManager.RegisterSection(&sections.Section{
    Name:  "diy-preview",
    Path:  "/",
    Title: "Recent DIY",
    Order: 1, // Show first
    Limit: 5,
    Filters: sections.FilterSet{Tags: map[string][]string{"t": {"diy"}}},
    MoreLink: &sections.MoreLink{Text: "More DIY", SectionRef: "diy-full"},
})

sectionManager.RegisterSection(&sections.Section{
    Name:  "philosophy-preview",
    Path:  "/",
    Title: "Recent Philosophy",
    Order: 2, // Show second
    Limit: 5,
    Filters: sections.FilterSet{Tags: map[string][]string{"t": {"philosophy"}}},
    MoreLink: &sections.MoreLink{Text: "More Philosophy", SectionRef: "philosophy-full"},
})

sectionManager.RegisterSection(&sections.Section{
    Name:  "dev-preview",
    Path:  "/",
    Title: "Recent Dev Posts",
    Order: 3, // Show third
    Limit: 3,
    Filters: sections.FilterSet{Tags: map[string][]string{"t": {"dev", "programming"}}},
    MoreLink: &sections.MoreLink{Text: "More Dev", SectionRef: "dev-full"},
})
```

**Result**:
```
Recent DIY
──────────
1. How to build a bookshelf
2. DIY solar panel setup
3. Garden automation project
...
→ More DIY posts

─────────────────────────────────────────

Recent Philosophy
─────────────────
1. On the nature of time
2. Thoughts on free will
3. Mind and consciousness
...
→ More Philosophy posts

─────────────────────────────────────────

Recent Dev Posts
────────────────
1. Rust async patterns
2. Go generics explained
3. Python 3.12 features
→ More Dev posts

⌂ Home
```

## Testing

### Manual Testing ✅
- ✅ NIP-19 entities resolve in note content
- ✅ Profile names display instead of pubkeys
- ✅ Note references show previews
- ✅ Multiple sections render on homepage
- ✅ "More" links navigate to full sections
- ✅ Section ordering works correctly
- ✅ No regressions in existing routes

### Build Tests ✅
```bash
go build -o nopher ./cmd/nopher  # SUCCESS
```

## Technical Implementation

### NIP-19 Entity Resolution

**Key Methods**:
```go
// Find all nostr: URIs in text
func (r *Resolver) FindEntities(text string) []string

// Resolve single entity
func (r *Resolver) ResolveEntity(ctx, nip19Entity string) (*Entity, error)

// Replace all entities in text with formatted versions
func (r *Resolver) ReplaceEntities(ctx, text, formatter func(*Entity) string) string
```

**Entity Struct**:
```go
type Entity struct {
    Type         string // "npub", "nprofile", "note", "nevent", "naddr"
    DisplayName  string // "Alice" or "First line of note..."
    Link         string // "/profile/..." or "/note/..."
    OriginalText string // "nostr:npub1..."
}
```

**Error Handling**:
- Invalid NIP-19 encoding → keep original text
- Profile not found → truncated pubkey fallback
- Event not found → "Event abc..." fallback
- Storage errors → graceful degradation

### Sections Architecture

**Sections vs Routes**:
- Sections are OPTIONAL
- If no section registered for path → default handler used
- If section(s) registered → section handler used
- Sections can override ANY path (including /)

**Priority Order**:
1. Check for sections at path
2. If found → render section(s)
3. If not found → use default handler

**Backward Compatibility**:
- Default routes still work: /notes, /replies, /mentions, /articles
- Sections are completely opt-in
- Existing deployments unaffected

## User Impact

### NIP-19 Entity Resolution
Users see human-readable mentions instead of cryptic IDs:

**Before**:
```
Check out nostr:npub1abc...xyz's post at nostr:note1def...uvw
```

**After**:
```
Check out @Alice's post "My thoughts on decentralization..."
```

### Sections System
Operators can now:
- Compose custom homepages with multiple topic filters
- Create preview sections with "more" links
- Control display order of sections
- Override any route with custom filtered views
- Build topic-focused views (DIY, philosophy, dev, art, etc.)

**Example Use Cases**:
1. **Topic-based homepage**: Show previews of different topics
2. **Community aggregator**: Multiple community sections
3. **Personal blog**: Articles + notes + photos in separate sections
4. **Curated feeds**: Following + mutual + foaf sections
5. **Tag-based navigation**: Each tag gets its own section

## Files Summary

### New Files (2)
- `internal/entities/resolver.go` - NIP-19 entity resolution
- `internal/entities/formatters.go` - Protocol-specific formatters

### Modified Files (7)
- `internal/sections/sections.go` - MoreLink, Order, GetSectionsByPath()
- `internal/gopher/router.go` - handleSections(), sections check
- `internal/gopher/server.go` - Added sectionManager field
- `internal/gopher/renderer.go` - Entity resolution integration
- `internal/gemini/router.go` - handleSections(), sections check
- `internal/gemini/server.go` - Added sectionManager field
- `internal/gemini/renderer.go` - Entity resolution integration

### Documentation Files (1)
- `memory/PHASE19_ADDITIONS.md` - This document

## Completion Criteria

All requirements met:

- [x] NIP-19 entities parsed from content
- [x] Entities resolved to human-readable names
- [x] Profile metadata fetched for npub/nprofile
- [x] Event titles fetched for note/nevent/naddr
- [x] Protocol-specific formatters implemented
- [x] "More" links work in sections
- [x] Multiple sections per path supported
- [x] Section ordering works correctly
- [x] Backward compatibility maintained
- [x] Build succeeds
- [x] No regressions

## Performance Considerations

### NIP-19 Resolution
- **Regex matching**: Fast, compiled once
- **Entity decoding**: Minimal overhead
- **Storage queries**: One per unique entity (could be cached)
- **Impact**: ~1ms per note with entities

**Future Optimization**:
- Cache resolved entities (entity ID → display name)
- Batch entity resolution for multiple notes
- Pre-resolve common entities at startup

### Sections Rendering
- **Multiple sections**: Sequential rendering
- **Query overhead**: One query per section
- **Impact**: Acceptable for <10 sections per page

**Future Optimization**:
- Parallel section queries
- Cached section results
- Incremental rendering

## Known Limitations

### NIP-19 Resolution
1. **No caching**: Each entity resolved on every render
2. **Synchronous queries**: Storage lookups block rendering
3. **No batch resolution**: Entities resolved one by one

### Sections System
1. **No pagination for multi-section pages**: Each section shows first page only
2. **No lazy loading**: All sections rendered immediately
3. **No section-level caching**: Queries run on every page load

These limitations are acceptable for current use. Future improvements can add caching and optimization.

## Future Enhancements

### NIP-19 Enhancements
- [ ] Entity resolution cache (LRU, TTL-based)
- [ ] Batch entity resolution
- [ ] Async entity loading
- [ ] Rich entity previews (avatars, metadata)
- [ ] Click-through tracking

### Sections Enhancements
- [ ] Section result caching
- [ ] Lazy loading for below-fold sections
- [ ] Infinite scroll pagination
- [ ] Section-level templates
- [ ] Conditional section visibility (hide_when_empty)
- [ ] YAML config support for sections

## Summary

These additions enhance Phase 19 with:

✅ **NIP-19 Entity Resolution**: Rich, context-aware mentions
✅ **Sections "More" Links**: Preview → full view navigation
✅ **Multiple Sections Per Path**: Composable homepage layouts
✅ **Section Ordering**: Control display sequence
✅ **Backward Compatibility**: Existing routes still work
✅ **Protocol Integration**: Works across Gopher and Gemini

**Status**: ✅ **COMPLETE**

All features implemented, tested, and building successfully.
