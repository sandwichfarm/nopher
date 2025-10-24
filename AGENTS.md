Agent Instructions for Nopher Development

Purpose

This document provides instructions for AI agents (like Claude) working on Nopher. It explains how to:
- Execute on phases and todo lists
- Interface with memory/ documentation
- Handle plan changes and keep documentation updated
- Maintain consistency across the project

Core Principles

1. **Memory is Source of Truth**
   - All design decisions are documented in memory/
   - Code implements what memory/ describes
   - When code and memory/ conflict, memory/ wins

2. **Phases are the Roadmap**
   - See memory/PHASES.md for implementation order
   - Each phase has clear deliverables and completion criteria
   - Phases can be worked independently where noted

3. **No Time Estimates**
   - We never provide time estimates
   - Focus on completion criteria, not duration
   - Work is done when criteria are met

4. **Document Changes**
   - When you discover something that changes the plan, update memory/
   - Keep memory/ in sync with implementation
   - Update PHASES.md if phase definition changes

5. **Documentation Philosophy - CRITICAL DISTINCTION**

   **📁 memory/ = Planning, Design, SDLC (Internal)**
   - ALL implementation planning documents
   - ALL phase completion tracking (PHASE*.md)
   - ALL design documents (architecture.md, storage_model.md, etc.)
   - ALL SDLC/project management files
   - ALL-CAPS documents about implementation status
   - Roadmaps, future plans, completion criteria
   - Technical specifications and rationale
   - FOR: Developers and AI agents working on implementation

   **📁 docs/ = User Guides (External)**
   - User-facing documentation ONLY
   - Deployment guides, configuration guides
   - Troubleshooting, API references
   - How-to guides for EXISTING features
   - Architecture overviews (conceptual, not status)
   - FOR: End users, operators, contributors using the software

   **RULE: If it mentions phases, completion status, or future implementation → memory/**
   **RULE: If it's a guide for using what exists now → docs/**

   See "Documentation Guidelines" section below for details

6. **Code Quality and Architecture**
   - Write modular, DRY (Don't Repeat Yourself) code
   - Keep files small and focused (<500 lines ideal, <1000 max)
   - Separate concerns into distinct packages
   - Prefer composition over monolithic implementations
   - See "Code Quality Standards" section below

Code Quality Standards

File Size and Modularity

❌ **AVOID: Huge monolithic files**
```
Bad:
internal/gopher/server.go (2000 lines)
  - Server setup
  - All route handlers
  - Gophermap generation
  - Text rendering
  - Caching logic
  - Query building
  - Error handling
```

✅ **PREFER: Small, focused modules**
```
Good:
internal/gopher/
  server.go (150 lines)      - Server setup, listener, main loop
  router.go (100 lines)      - Route matching and dispatch
  handlers.go (200 lines)    - Route handler functions
  gophermap.go (150 lines)   - Gophermap generation
  renderer.go (200 lines)    - Text rendering for events
  sections.go (150 lines)    - Section queries
  cache.go (100 lines)       - Gopher-specific caching
  types.go (50 lines)        - Gopher type definitions
```

**Guidelines:**
- **Target file size: 100-300 lines** (code, not comments)
- **Maximum file size: 500 lines** (hard limit)
- **If approaching 500 lines:** Split into multiple files
- **One responsibility per file:** Server, routing, rendering, etc.
- **Clear naming:** File name should describe its single purpose

DRY (Don't Repeat Yourself)

❌ **AVOID: Copy-pasted code**
```go
// Bad: Same logic repeated in three places
func (g *GopherServer) handleNotes() {
    events := g.queryEvents(kindNotes)
    for _, e := range events {
        // 20 lines of rendering logic
    }
}

func (g *GopherServer) handleArticles() {
    events := g.queryEvents(kindArticles)
    for _, e := range events {
        // Same 20 lines of rendering logic (copy-pasted!)
    }
}

func (g *GopherServer) handleInbox() {
    events := g.queryEvents(kindInbox)
    for _, e := range events {
        // Same 20 lines of rendering logic (copy-pasted again!)
    }
}
```

✅ **PREFER: Shared, reusable functions**
```go
// Good: Extract common logic
func (g *GopherServer) renderEvents(events []*Event) string {
    // 20 lines of rendering logic (written once)
}

func (g *GopherServer) handleNotes() {
    events := g.queryEvents(kindNotes)
    return g.renderEvents(events)
}

func (g *GopherServer) handleArticles() {
    events := g.queryEvents(kindArticles)
    return g.renderEvents(events)
}

func (g *GopherServer) handleInbox() {
    events := g.queryEvents(kindInbox)
    return g.renderEvents(events)
}
```

**Rules:**
- **If code appears twice:** Extract to a function
- **If code appears 3+ times:** Extract to a shared package
- **Common utilities:** Put in internal/common or pkg/
- **Protocol-specific:** Keep in protocol package but extract functions

Package Organization

❌ **AVOID: Flat structure or God packages**
```
Bad (flat):
internal/
  everything.go (5000 lines)
  utils.go (2000 lines)
  helpers.go (1500 lines)

Bad (God package):
internal/server/
  server.go (contains Gopher, Gemini, Finger all mixed together)
```

✅ **PREFER: Clear package boundaries**
```
Good:
internal/
  config/           - Configuration loading and validation
    config.go
    env.go
    validation.go
  storage/          - Storage layer and Khatru integration
    storage.go
    khatru.go
    migrations.go
  nostr/            - Nostr client and relay communication
    client.go
    relay.go
    discovery.go
  sync/             - Event synchronization engine
    engine.go
    filters.go
    graph.go
  markdown/         - Markdown parsing and conversion
    parser.go
    gopher.go
    gemini.go
  gopher/           - Gopher protocol server (small files!)
    server.go
    router.go
    handlers.go
    gophermap.go
  gemini/           - Gemini protocol server
    server.go
    router.go
    handlers.go
    gemtext.go
  finger/           - Finger protocol server
    server.go
    query.go
    renderer.go
  cache/            - Caching layer
    cache.go
    memory.go
    redis.go
  common/           - Shared utilities
    helpers.go
    validators.go
```

**Package Guidelines:**
- **One package per major feature**
- **Package name describes its purpose**
- **Packages don't import their siblings** (avoid circular deps)
- **internal/ for private code, pkg/ for public APIs**
- **Keep packages small:** 5-10 files max per package

Function Size

❌ **AVOID: Giant functions**
```go
// Bad: 200-line function doing everything
func (s *SyncEngine) Sync() error {
    // 50 lines: connect to relays
    // 50 lines: build filters
    // 50 lines: process events
    // 50 lines: update database
    // This is unreadable and untestable!
}
```

✅ **PREFER: Small, focused functions**
```go
// Good: Multiple small functions
func (s *SyncEngine) Sync() error {
    relays, err := s.connectToRelays()
    if err != nil {
        return err
    }
    defer s.disconnectRelays(relays)

    filters := s.buildFilters()
    events := s.subscribeAndCollect(relays, filters)
    return s.processEvents(events)
}

func (s *SyncEngine) connectToRelays() ([]*Relay, error) {
    // 20 lines: focused on connection logic
}

func (s *SyncEngine) buildFilters() []Filter {
    // 15 lines: focused on filter building
}

func (s *SyncEngine) subscribeAndCollect(relays []*Relay, filters []Filter) []*Event {
    // 30 lines: focused on event collection
}

func (s *SyncEngine) processEvents(events []*Event) error {
    // 25 lines: focused on processing
}
```

**Function Guidelines:**
- **Target: 10-30 lines per function**
- **Maximum: 50 lines** (hard limit)
- **One responsibility:** Function name should describe exactly what it does
- **If function has 'and' in description:** Split it
- **Extract nested loops:** Put inner loop in separate function

Separation of Concerns

❌ **AVOID: Mixed responsibilities**
```go
// Bad: Server handles everything
type GopherServer struct {
    listener net.Listener
    db       *Database        // Server shouldn't know about DB!
    cache    map[string]string // Server shouldn't manage cache!
    markdown *MarkdownParser   // Server shouldn't parse markdown!
}

func (s *GopherServer) handleEvent(id string) {
    // Queries database directly
    row := s.db.Query("SELECT * FROM events WHERE id = ?", id)

    // Parses markdown
    html := s.markdown.Parse(row.Content)

    // Renders gophermap
    // Manages cache
    // Everything in one place!
}
```

✅ **PREFER: Clear separation with interfaces**
```go
// Good: Server depends on abstractions
type GopherServer struct {
    listener net.Listener
    storage  EventStorage      // Interface, not concrete type
    renderer Renderer          // Interface for rendering
    cache    Cache             // Interface for caching
}

type EventStorage interface {
    GetEvent(id string) (*Event, error)
    QueryEvents(filter Filter) ([]*Event, error)
}

type Renderer interface {
    RenderEvent(event *Event) (string, error)
    RenderGophermap(events []*Event) (string, error)
}

type Cache interface {
    Get(key string) (string, bool)
    Set(key, value string, ttl time.Duration)
}

func (s *GopherServer) handleEvent(id string) {
    // Uses interface methods
    event, err := s.storage.GetEvent(id)
    if err != nil {
        return s.errorResponse(err)
    }

    rendered, err := s.renderer.RenderEvent(event)
    if err != nil {
        return s.errorResponse(err)
    }

    s.cache.Set(id, rendered, 5*time.Minute)
    return rendered
}
```

**Separation Guidelines:**
- **Server:** Handles protocol, routing, HTTP/TCP only
- **Storage:** Handles database, queries, persistence
- **Renderer:** Handles content transformation
- **Cache:** Handles caching logic
- **Each layer has clear interface**
- **Dependencies point inward** (server → storage, not storage → server)

When to Extract

Extract to a new function when:
- ✅ Code is repeated (DRY violation)
- ✅ Function exceeds 50 lines
- ✅ Nested loops (put inner loop in function)
- ✅ Try-catch/error handling blocks are large
- ✅ Comment says "now we..." (each "now" is a function)

Extract to a new file when:
- ✅ File exceeds 500 lines
- ✅ File handles multiple concerns
- ✅ Scrolling required to see full type definition
- ✅ File name contains "and" (split it)

Extract to a new package when:
- ✅ Files share common prefix (gopher_server.go, gopher_render.go → package gopher)
- ✅ Functionality is cohesive and reusable
- ✅ 10+ files in one package
- ✅ Clear domain boundary exists

Practical Example: Refactoring

**Before: Monolithic**
```
internal/server/server.go (1500 lines)
```

**After: Modular**
```
internal/
  gopher/
    server.go (150 lines)
    router.go (100 lines)
    handlers.go (200 lines)
    gophermap.go (150 lines)
    renderer.go (200 lines)
  gemini/
    server.go (150 lines)
    router.go (100 lines)
    handlers.go (180 lines)
    gemtext.go (150 lines)
  finger/
    server.go (120 lines)
    query.go (80 lines)
    renderer.go (100 lines)
```

**Refactoring checklist:**
1. ✅ Each protocol has own package
2. ✅ No file exceeds 300 lines
3. ✅ Each file has single responsibility
4. ✅ No duplicate code between packages
5. ✅ Clear interfaces between components

Testing Small Modules

Small, modular code is easier to test:

```go
// Easy to test: Small, pure function
func escapeSelector(s string) string {
    return strings.ReplaceAll(s, "\t", "%09")
}

// Test is simple
func TestEscapeSelector(t *testing.T) {
    assert.Equal(t, "foo%09bar", escapeSelector("foo\tbar"))
}

// Easy to test: Function with interface dependency
func RenderEvent(event *Event, renderer Renderer) (string, error) {
    return renderer.Render(event)
}

// Test uses mock
func TestRenderEvent(t *testing.T) {
    mock := &MockRenderer{result: "rendered"}
    output, _ := RenderEvent(&Event{}, mock)
    assert.Equal(t, "rendered", output)
}
```

Hard to test: 500-line function with embedded DB queries, HTTP calls, file I/O, and complex logic all mixed together.

Red Flags in Code Review

Watch for these anti-patterns:

🚩 **"God Object"** - One struct/type does everything
🚩 **"Blob File"** - File over 1000 lines
🚩 **"Copy-Paste Programming"** - Same code in multiple places
🚩 **"Lasagna Code"** - Too many layers of indirection (also bad!)
🚩 **"Yo-yo Problem"** - Navigate up and down inheritance/delegation
🚩 **"Swiss Army Knife"** - Function/package tries to do everything
🚩 **"Magic Numbers"** - Unexplained constants (use named constants)
🚩 **"Premature Optimization"** - Complex code for unclear performance gain

If you see these, refactor immediately.

Summary: Code Quality Checklist

Before committing code, verify:

✅ **Files:**
- [ ] No file exceeds 500 lines
- [ ] Each file has single, clear purpose
- [ ] File name describes its contents

✅ **Functions:**
- [ ] No function exceeds 50 lines
- [ ] Each function does one thing
- [ ] Function names are descriptive

✅ **DRY:**
- [ ] No code duplicated
- [ ] Common logic extracted
- [ ] Utilities in shared packages

✅ **Packages:**
- [ ] Clear package boundaries
- [ ] Packages have focused purpose
- [ ] No circular dependencies

✅ **Separation:**
- [ ] Concerns are separated
- [ ] Interfaces define contracts
- [ ] Dependencies point inward

✅ **Tests:**
- [ ] Small functions are tested
- [ ] Tests are readable
- [ ] Mocks used for dependencies

**When in doubt: Smaller is better.**

Working with Memory

Reading Memory

Before starting any work:

1. **Read memory/README.md** - Overview of the project
2. **Read memory/PHASES.md** - Find which phase you're working on
3. **Read relevant memory/*.md files** - Understand the design

Example: Starting Phase 7 (Gopher Server)
```
Read these files:
- memory/PHASES.md → Find Phase 7 details
- memory/architecture.md → Understand overall design
- memory/ui_export.md → Gopher rendering specs
- memory/markdown_conversion.md → How to convert content
- memory/storage_model.md → How to query Khatru
- memory/glossary.md → Terminology reference
```

When to Read Memory

- **Starting a new phase** → Read all related memory/ files
- **Implementing a feature** → Read specific design docs
- **Debugging an issue** → Check if memory/ explains expected behavior
- **Making a decision** → Check if memory/ has guidance
- **User asks "how should X work?"** → Check memory/ first

Updating Memory

When to Update

Update memory/ documents when:

✅ **Implementation reveals design flaws**
   - You discover the design won't work as described
   - Example: "NIP-65 parsing needs additional field"
   - Action: Update memory/relay_discovery.md with new field

✅ **Requirements change**
   - User requests a change to the plan
   - Example: "Add support for NIP-42 authentication"
   - Action: Update relevant memory/ files and PHASES.md

✅ **Better approach discovered**
   - You find a superior implementation method
   - Example: "Caching strategy should use LRU instead of TTL-only"
   - Action: Update memory/caching.md with new strategy

✅ **Missing information**
   - Memory/ is incomplete or unclear
   - Example: "No spec for error handling in sync engine"
   - Action: Add error handling section to memory/sync_scope.md

❌ **Do NOT update for:**
   - Minor implementation details (variable names, file paths)
   - Temporary debugging notes
   - Personal preferences on code style
   - Things already documented correctly

How to Update

1. **Read the current document fully** before editing
2. **Make surgical changes** - only update what needs changing
3. **Preserve existing information** unless it's wrong
4. **Keep the same format and style** as existing docs
5. **Update references** if you change terminology
6. **Don't remove sections** unless they're truly obsolete

Example Update Workflow

Scenario: Implementing Phase 4 (Sync Engine), you discover NIP-65 relay hints need a "last_verified" field.

Steps:
1. Note the issue: relay hints can become stale
2. Check memory/storage_model.md: no "last_verified" field exists
3. Check memory/relay_discovery.md: no mention of verification
4. Edit memory/storage_model.md:
   - Add "last_verified INTEGER" to relay_hints table
   - Update purpose: "Track freshness and verification status"
5. Edit memory/relay_discovery.md:
   - Add section on relay verification
   - Explain when last_verified is updated
6. Continue implementing with new field

Updating PHASES.md

Only update PHASES.md when:

✅ **Phase definition changes**
   - Deliverables change significantly
   - Completion criteria change
   - Dependencies change

✅ **New phase needed**
   - User requests major new feature
   - Implementation requires new phase

✅ **Phase becomes obsolete**
   - Technology change makes phase unnecessary
   - Example: If we switched from Khatru, Phase 2 would change

❌ **Do NOT update PHASES.md for:**
   - Minor tweaks to implementation
   - File name changes
   - Small additions to deliverables

Format for Phase Updates

When updating a phase:
```markdown
Phase X: <Name>

Goal: <One sentence goal>

Deliverables:
- <Concrete deliverable 1>
- <Concrete deliverable 2>

Completion Criteria:
- <Testable criterion 1>
- <Testable criterion 2>

Dependencies: Phase Y, Phase Z

Files to Create:
- <path/to/file.go>

Memory Updates:
- <document.md> (if <condition>)
```

Working with Todo Lists

Todo lists are managed by the user or project lead. As an agent:

**Reading Todos**
- If a todo list exists (GitHub Issues, project board, etc.), read it
- Understand which tasks are assigned and prioritized
- Ask for clarification if a todo is ambiguous

**Updating Todos**
- Mark todos complete when criteria are met
- Add new todos if you discover missing work
- Break down large todos into smaller ones
- Flag blockers immediately

**Todo Format** (if creating):
```markdown
## Phase X: <Phase Name>

- [ ] Task 1: <Description>
- [ ] Task 2: <Description>
- [x] Task 3: <Completed task>
```

Handling Plan Changes

User Requests a Change

Example: User says "We need to support PostgreSQL after all"

Your response process:

1. **Acknowledge the change**
   "I'll update the plan to add PostgreSQL support."

2. **Identify impacted documents**
   - memory/storage_model.md (backend choice)
   - memory/configuration.md (postgres config)
   - memory/architecture.md (database section)
   - memory/PHASES.md (Phase 2 deliverables)

3. **Update documents in order**
   - Start with most fundamental (storage_model.md)
   - Then configuration (configuration.md)
   - Then overview (architecture.md)
   - Finally roadmap (PHASES.md)

4. **Summarize changes**
   "Updated 4 documents to add PostgreSQL:
   - storage_model.md: Added Postgres schema
   - configuration.md: Added postgres_url config
   - architecture.md: Updated backend options
   - PHASES.md: Phase 2 now includes Postgres setup"

5. **Proceed with implementation**

You Discover a Design Issue

Example: While implementing Phase 7, you realize Gopher selectors need escaping.

Your process:

1. **Document the issue clearly**
   "Gopher selectors can contain tab characters which break the protocol. Need escaping."

2. **Check if memory/ addresses it**
   - Read memory/ui_export.md Gopher section
   - No mention of selector escaping

3. **Propose solution**
   "I'll add selector escaping to the Gopher renderer. This requires updating memory/ui_export.md."

4. **Update memory/ first**
   - Edit memory/ui_export.md
   - Add "Selector Escaping" section under Gopher Server
   - Document escape rules (tabs → %09, etc.)

5. **Implement the solution**
   - Add escaping function to internal/gopher/selector.go
   - Add tests

6. **Confirm alignment**
   "Implementation matches updated memory/ui_export.md."

Technology Change

Example: User decides to switch from Khatru to custom storage.

Your response:

1. **Assess impact**
   "This is a major change affecting Phase 2, 4, 5, and core architecture."

2. **List affected documents**
   - memory/architecture.md (core design)
   - memory/storage_model.md (complete rewrite)
   - memory/PHASES.md (Phase 2 complete rewrite)
   - memory/README.md (overview mentions Khatru)

3. **Recommend approach**
   "This requires significant documentation updates. Recommend:
   1. Update architecture.md with new storage design
   2. Rewrite storage_model.md for custom implementation
   3. Update PHASES.md Phase 2 with new deliverables
   4. Update README.md to remove Khatru references"

4. **Execute updates** (with user approval)

5. **Flag dependencies**
   "Note: This delays Phase 2 as we need to implement event storage from scratch."

Workflow Patterns

Starting a New Phase

1. **Read PHASES.md** → Find your phase
2. **Check dependencies** → Verify previous phases complete
3. **Read memory/ docs** → Listed in "Memory Updates" section
4. **Create todo checklist** → From "Deliverables"
5. **Implement** → Following memory/ specs
6. **Test** → Meet "Completion Criteria"
7. **Update memory/** → If anything changed
8. **Mark phase complete**

Implementing a Feature

1. **Find in PHASES.md** → Which phase is this?
2. **Read design docs** → Understand expected behavior
3. **Write tests first** → Based on completion criteria
4. **Implement** → Make tests pass
5. **Check memory/ alignment** → Does code match design?
6. **Update memory/** → If design needed changes
7. **Mark complete**

Fixing a Bug

1. **Understand expected behavior** → Check memory/
2. **Identify root cause** → Debug
3. **Check if design is wrong** → Is memory/ incorrect?
   - If yes: Update memory/ first, then fix code
   - If no: Fix code to match memory/
4. **Add regression test**
5. **Verify fix**

Adding Documentation

1. **Determine type**:
   - Design decision → Update memory/
   - User guide → Add to docs/
   - Code comment → Add inline
   - API reference → Generate from code

2. **For memory/ updates**:
   - Find appropriate existing document
   - Add section maintaining current style
   - Update references if needed

3. **For new memory/ documents**:
   - Rare! Usually use existing docs
   - If truly needed: match format of existing docs
   - Add to memory/README.md references

Communication Guidelines

When Asking for Clarification

Good:
- "memory/sync_scope.md mentions 'mutual follows' but doesn't define the query. Should it be bidirectional kind-3 check?"
- "Phase 7 deliverable 'archive navigation' is unclear. Should this be by year/month like memory/layouts_sections.md suggests?"

Bad:
- "What should I do?" (too vague)
- "This doesn't work" (no context)

When Reporting Progress

Good:
- "Completed Phase 7 deliverables:
   ✓ TCP server on port 70
   ✓ Selector routing
   ✓ Gophermap generation
   ⏳ Thread display (in progress)
   - Event querying (not started)"

Bad:
- "Made progress" (not measurable)
- "Almost done" (no specifics)

When Proposing Changes

Good:
- "Current design in memory/caching.md uses TTL-only expiration. Propose adding LRU eviction for memory-constrained environments. This requires updating caching.md section 'Eviction Strategy' and adding 'max_items' to config."

Bad:
- "Should change caching" (no reasoning)
- "This way is better" (no justification)

Documentation Guidelines

⚠️ **CRITICAL: File Location Matters** ⚠️

**memory/ vs docs/ - WRONG vs RIGHT:**

❌ **WRONG - These belong in memory/ NOT docs/:**
```
docs/PHASE*_COMPLETION.md      ← NO! Phase tracking = memory/
docs/[DESIGN_TOPIC].md         ← NO! If it's design/planning = memory/
```

✅ **RIGHT - Correct locations:**
```
memory/PHASE*_COMPLETION.md    ← YES! Phase completion docs go here
memory/[design_topic].md       ← YES! Implementation design goes here
docs/[feature]-guide.md        ← YES! User guide for features
docs/installation.md           ← YES! How to install what exists NOW
```

**📁 docs/ - User-Facing Documentation**

**PURPOSE:** Help users/operators/contributors USE the software that exists today.

**Applies to:**
- README.md (project overview and quick start)
- CONTRIBUTING.md (contribution guidelines)
- docs/ directory (user guides only)

**What BELONGS in docs/:**
- ✅ Deployment guides for current functionality
- ✅ Configuration reference for working options
- ✅ Troubleshooting real issues
- ✅ API reference for implemented endpoints
- ✅ How-to guides for features that work NOW
- ✅ Architecture overview (conceptual, for users)

**What NEVER goes in docs/:**
- ❌ PHASE*.md files (phase completion tracking)
- ❌ "Phase X Complete" status updates
- ❌ Implementation roadmaps or future plans
- ❌ "Next steps" or "Coming soon" sections
- ❌ Completion checklists or criteria
- ❌ SDLC/project management content
- ❌ Technical design rationale
- ❌ ALL-CAPS status documents

**📁 memory/ - Planning, Design, SDLC Documentation**

**PURPOSE:** Track implementation, design decisions, and project planning.

**Applies to:**
- memory/PHASES.md (implementation roadmap)
- memory/PHASE*_COMPLETION.md (phase tracking)
- memory/*.md (all design documents)

**What BELONGS in memory/:**
- ✅ ALL PHASE*.md files (completion tracking)
- ✅ Implementation roadmaps (PHASES.md)
- ✅ Technical design documents
- ✅ Architecture decisions and rationale
- ✅ Future plans and feature specs
- ✅ Completion criteria for phases
- ✅ SDLC tracking and status
- ✅ Configuration schema design
- ✅ ALL-CAPS documents about status

**Examples of what goes in memory/:**
- ✅ PHASE*_COMPLETION.md files
- ✅ Technical design documents
- ✅ Implementation specifications
- ✅ Architecture decision records

Example README Status Section

**Good:**
```markdown
## Status

⚠️ **Early Development** - Not yet ready for production use.

Current implementation status:
- Configuration system with YAML parsing and validation
- Storage layer with Khatru integration and SQLite backend
- Custom tables for relay hints, social graph, sync state, and aggregates

Protocol servers (Gopher, Gemini, Finger) are not yet implemented.
```

**Bad:**
```markdown
## Status

🚧 **Phase 2 Complete - Storage Layer**

The project is progressing well:
- ✅ Phase 0: Bootstrap complete
- ✅ Phase 1: Configuration system complete
- ✅ Phase 2: Storage layer complete
- ⏳ Phase 3: Relay discovery (next)
- 🔲 Phase 4-16: Planned

Next: Phase 3 (Nostr Client and Relay Discovery)

See `memory/PHASES.md` for the complete implementation roadmap.
```

Why? The "good" example tells users what works RIGHT NOW. The "bad" example is project management tracking that belongs in memory/PHASES.md, not user-facing docs.

Quick Reference

Before Every Task
1. ✅ Read memory/PHASES.md → Find your phase
2. ✅ Read related memory/ docs → Understand design
3. ✅ Check dependencies → Previous phases complete?

During Implementation
1. ✅ Follow memory/ specs → Code matches design
2. ✅ Write tests → Meet completion criteria
3. ✅ Note discrepancies → Design vs reality

After Completion
1. ✅ Update memory/ → If design changed
2. ✅ Mark deliverables complete → In todos/PHASES.md
3. ✅ Document what changed → For next phase

When Stuck
1. ✅ Re-read memory/ docs → Might have missed something
2. ✅ Check glossary → Terminology confusion?
3. ✅ Ask specific question → Reference docs/phases
4. ✅ Propose solution → With memory/ updates

Memory Document Map

By Phase:

Phase 0 (Bootstrap):
- cicd.md, distribution.md

Phase 1 (Config):
- configuration.md

Phase 2 (Storage):
- storage_model.md, architecture.md

Phase 3 (Relay Discovery):
- relay_discovery.md, sequence_seed_discovery_sync.md, nips.md

Phase 4 (Sync Engine):
- sync_scope.md, sequence_seed_discovery_sync.md, operations.md

Phase 5 (Aggregates):
- inbox_outbox.md, nips.md

Phase 6 (Markdown):
- markdown_conversion.md

Phase 7 (Gopher):
- ui_export.md, glossary.md (Gopher section)

Phase 8 (Gemini):
- ui_export.md, glossary.md (Gemini section)

Phase 9 (Finger):
- ui_export.md, glossary.md (Finger section)

Phase 10 (Caching):
- caching.md, cache_keys_hashing.md, sequence_cache_live.md

Phase 11 (Sections):
- layouts_sections.md

Phase 12 (Operations):
- operations.md, diagnostics.md

Phase 13 (Publisher):
- inbox_outbox.md, nips.md

Phase 14 (Security):
- security_privacy.md

Phase 15 (Testing):
- testing.md

Phase 16 (Distribution):
- distribution.md, cicd.md

By Topic:

Architecture & Design:
- README.md, architecture.md, glossary.md

Nostr Integration:
- nips.md, relay_discovery.md, inbox_outbox.md, sync_scope.md

Protocols:
- ui_export.md, markdown_conversion.md

Storage & Caching:
- storage_model.md, caching.md, cache_keys_hashing.md

Operations:
- operations.md, diagnostics.md, security_privacy.md

Distribution:
- distribution.md, cicd.md

All Documents:
- Run: ls memory/*.md

Example Sessions

Session 1: Starting Phase 0

Agent: "Starting Phase 0 (Project Bootstrap). Reading memory/PHASES.md..."

Agent: "Phase 0 requires:
- Go module initialization
- Directory structure
- CI/CD setup

Reading memory/cicd.md and memory/distribution.md for specifications..."

Agent: "Creating project structure:
- go mod init github.com/user/nopher
- mkdir -p cmd/nopher internal pkg scripts .github/workflows
- Creating Makefile from memory/cicd.md examples
- Creating GitHub Actions workflows from memory/cicd.md specs"

Agent: "Phase 0 completion criteria:
✓ make build produces binary
✓ make test runs
✓ make lint passes
✓ GitHub Actions workflows functional
✓ Can create snapshot release

All criteria met. Phase 0 complete."

Session 2: Implementing Phase 7 with Discovery

Agent: "Starting Phase 7 (Gopher Server). Reading memory/ui_export.md..."

Agent: "Working on selector routing. Memory/ui_export.md shows:
- Selector '/' → main menu
- Selector '/notes' → notes section
- Selector '/event/<id>' → individual event

But doesn't specify how to handle invalid selectors."

Agent: "Checking memory/glossary.md for Gopher error handling..."

Agent: "Found: Gopher item type '3' is for errors. Will return type 3 for invalid selectors. Updating memory/ui_export.md to document this..."

Agent: "Edit: memory/ui_export.md
Added section 'Error Handling' under Gopher Server:
- Invalid selectors return type '3' (error)
- Error message explains valid selectors"

Agent: "Implementing selector routing with error handling per updated memory/ui_export.md..."

Agent: "Tests pass. Gopher server handles invalid selectors correctly."

Session 3: Handling User Change Request

User: "Actually, we should support Redis caching in addition to in-memory."

Agent: "Understood. This affects Phase 10 (Caching Layer). Let me update the plan."

Agent: "Reading current memory/caching.md..."

Agent: "Current design mentions Redis as 'optional' but no implementation details. Will add Redis backend specifications."

Agent: "Updating documents:
1. memory/caching.md → Add Redis backend section
2. memory/configuration.md → Add Redis config options
3. memory/PHASES.md → Update Phase 10 deliverables"

Agent: "Updated memory/caching.md:
- Added 'Redis Backend' section with connection pooling
- Added configuration: redis_url, redis_max_connections
- Documented key namespacing strategy"

Agent: "Updated memory/configuration.md:
- Added caching.redis_url config option
- Added caching.redis_max_connections option
- Updated example config"

Agent: "Updated memory/PHASES.md:
- Phase 10 deliverables now include: 'Redis cache implementation'
- Added file: internal/cache/redis.go"

Agent: "All documents updated. Ready to implement Redis caching in Phase 10."

Session 4: Discovering Design Flaw

Agent: "Implementing NIP-10 thread resolution in Phase 5..."

Agent: "Issue discovered: Events can have multiple 'e' tags (mentions, not just replies). Current memory/inbox_outbox.md assumes all 'e' tags are thread relationships."

Agent: "Checking NIP-10 spec... NIP-10 uses tag positions and markers ('root', 'reply', 'mention') to differentiate."

Agent: "Current memory/inbox_outbox.md doesn't account for marked tags. This is a design gap."

Agent: "Updating memory/inbox_outbox.md:
- Adding section: 'NIP-10 Tag Interpretation'
- Documenting: Use 'reply' marker for threads, ignore 'mention'
- Fallback: Last 'e' tag is reply if no markers"

Agent: "Also updating memory/glossary.md:
- Expanding NIP-10 explanation with marker types"

Agent: "Continuing implementation with corrected NIP-10 interpretation..."

Summary

As an agent working on Nopher:

1. **Always read memory/ before implementing**
2. **Update memory/ when you discover gaps or changes**
3. **Keep PHASES.md synchronized with reality**
4. **Document changes clearly**
5. **Never provide time estimates**
6. **Focus on completion criteria, not speed**
7. **Ask questions referencing specific memory/ docs**
8. **Communicate progress with concrete deliverables**

Memory is the source of truth. Code implements memory. Changes update memory first, then code.
