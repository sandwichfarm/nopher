Composable Layouts and Sections

Concept
- Sections are query-driven views over indexed data. Pages are composed of sections.
- Each section defines what to show (filters) and how to show it (template), plus archives and feeds.

Section fields
- id: stable identifier
- title: display name
- source: inbox|outbox|all
- filter: structured selectors (see Filter spec) describing kinds/authors/relations/tags/time
- transform: grouping or threading options (group_by_thread, collapse_reposts)
- sort: field and order (e.g., -created_at)
- limit: max items per page; page_size for pagination
- template: list|cards|threaded|gallery|table (pluggable)
- archive: by month|year with route pattern
- feeds: enable rss/json
- hide_when_empty: true|false

Filter spec (examples)
- kinds: [1, 30023, 7, 9735]
- authors: [owner] | following | mutual | foaf:2 | allowlist:[npub...] | denylist:[...]
- is_reply: true|false
- thread_of: <event id>
- replies_to: owner|section:<id>
- mentions: owner | [npub...]
- p_tags: [npub...]
- e_tags: [<event id>...]
- hashtag: ["art", "dev"]
- d_tag: ["slug"]
- lang: ["en"]
- has_reactions: true|false
- min_zap_sats: N
- reaction_chars: ["+"]
- since/until: timestamps; or last_days: N
- include_direct_mentions: true|false
- include_threads_of_mine: true|false

Default layout (when none configured)

Homepage (/ or empty selector):
- **Default behavior**: Auto-generated menu (gophermap/gemtext) with links to all sections
- **Fully customizable**: Configure via pages.home.layout or create a section with path: "/"
- **Composable**: Can show multiple sections, single section, or custom content

Default sections (user-facing paths):
- /notes - Owner's notes (outbox, kinds:[1], is_reply:false, limit:20)
- /articles - Owner's articles (outbox, kinds:[30023], limit:10)
- /replies - Replies to owner (inbox, filter: replies_to:owner)
- /mentions - Mentions of owner (inbox, filter: mentions:owner)
- /archive - Time-based archives (by year/month)
- /about - Owner profile (kind 0)
- /diagnostics - System status

Customization Options for Homepage (/):
1. **Menu (default)**: Auto-generated links to all sections
2. **Composed page**: Multiple sections via pages.home.layout (e.g., profile + notes + replies)
3. **Single section**: Create a section with path: "/" (e.g., just show recent notes)
4. **Custom template**: Use any template (list, threaded, cards, etc.)

Note: "inbox" and "outbox" are internal source identifiers in section config, not exposed as paths.
Sections use source: "inbox" or source: "outbox" internally, but are accessed via descriptive paths like /replies or /notes.

Example (YAML fragment)

layout:
  sections:
    about:
      id: "about"
      path: "/about"                  # User-facing path
      title: "About"
      source: "outbox"                # Internal: where data comes from
      filter: { kinds: [0], authors: [owner] }
      template: "profile"
      hide_when_empty: false
    notes:
      id: "notes"
      path: "/notes"                  # User-facing path
      title: "Notes"
      source: "outbox"                # Internal: owner's content
      filter: { kinds: [1], authors: [owner], is_reply: false }
      sort: "-created_at"
      limit: 20
      template: "list"
      archive: { by: "month", route: "/archive/notes/{YYYY}/{MM}" }
    articles:
      id: "articles"
      path: "/articles"               # User-facing path
      title: "Articles"
      source: "outbox"                # Internal: owner's content
      filter: { kinds: [30023], authors: [owner] }
      sort: "-created_at"
      limit: 10
      template: "list"
      archive: { by: "year", route: "/archive/articles/{YYYY}" }
    replies:
      id: "replies"
      path: "/replies"                # User-facing path
      title: "Replies"
      source: "inbox"                 # Internal: content targeting owner
      filter: { kinds: [1], replies_to: "owner" }
      sort: "-created_at"
      limit: 50
      template: "threaded"
    mentions:
      id: "mentions"
      path: "/mentions"               # User-facing path
      title: "Mentions"
      source: "inbox"                 # Internal: content mentioning owner
      filter: { mentions: "owner" }
      sort: "-created_at"
      limit: 50
      template: "list"
  pages:
    home:
      path: "/"                       # Root path (customizable)
      layout:
        - ["about"]                   # Row 1: profile section
        - ["notes", "articles"]       # Row 2: two sections side-by-side
        - ["replies", "mentions"]     # Row 3: two sections side-by-side

# Alternative: Custom homepage (just recent notes)
# pages:
#   home:
#     path: "/"
#     layout:
#       - ["notes"]

# Alternative: Custom homepage (single custom section at /)
# sections:
#   homepage:
#     path: "/"                       # Override default menu
#     title: "Welcome"
#     source: "outbox"
#     filter: { kinds: [1, 30023], authors: [owner] }
#     sort: "-created_at"
#     limit: 10
#     template: "list"

Notes
- The homepage (/) is fully configurable via pages.home.layout or by creating a section with path: "/"
- Default behavior: Auto-generated menu linking to all sections
- Custom behavior: Compose any sections you want, or create a single section at /
- Sections obey global sync scope by default; they can narrow to authors:[owner] etc.
- Archives are optional per section; feeds can be generated per section.
- Templates are themeable; operators can choose different templates per section.
