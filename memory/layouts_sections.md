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
- Profile header (owner's kind 0)
- Recent Notes: outbox, kinds:[1], is_reply:false, limit:20, template:list
- Recent Comments: outbox, kinds:[1], is_reply:true, template:threaded, limit:20
- Recent Articles: outbox, kinds:[30023], limit:10, template:list
- Archives: month/year pages for Notes and Articles

Example (YAML fragment)

layout:
  sections:
    profile:
      id: "profile"
      title: "Profile"
      source: "outbox"
      filter: { kinds: [0], authors: [owner] }
      template: "profile"
      hide_when_empty: false
    notes:
      id: "notes"
      title: "Recent Notes"
      source: "outbox"
      filter: { kinds: [1], authors: [owner], is_reply: false }
      sort: "-created_at"
      limit: 20
      template: "list"
      archive: { by: "month", route: "/archive/notes/{YYYY}/{MM}" }
      feeds: { rss: true, json: true }
    comments:
      id: "comments"
      title: "Recent Comments"
      source: "outbox"
      filter: { kinds: [1], authors: [owner], is_reply: true }
      sort: "-created_at"
      limit: 20
      template: "threaded"
    articles:
      id: "articles"
      title: "Articles"
      source: "outbox"
      filter: { kinds: [30023], authors: [owner] }
      sort: "-created_at"
      limit: 10
      template: "list"
      archive: { by: "year", route: "/archive/articles/{YYYY}" }
    interactions:
      id: "interactions"
      title: "Interactions"
      source: "inbox"
      filter: { kinds: [7, 9735], mentions: "owner" }
      sort: "-created_at"
      limit: 50
      template: "list"
  pages:
    home:
      layout:
        - ["profile"]
        - ["notes"]
        - ["comments", "articles"]
        - ["interactions"]

Notes
- Sections obey global sync scope by default; they can narrow to authors:[owner] etc.
- Archives are optional per section; feeds can be generated per section.
- Templates are themeable; operators can choose different templates per section.
