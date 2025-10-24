Protocol Rendering

Gopher Server (RFC 1436)
- Serves on port 70; responds to selectors with gophermaps or text files.
- Homepage (selector "/" or empty): fully configurable via layout.pages.home or layout.sections
  - Default: Auto-generated gophermap menu with links to all sections
  - Customizable: Can show composed sections, single section, or custom content
- Default sections (all configurable via layout.sections):
  - /notes - Owner's notes (kind 1, non-replies)
  - /articles - Owner's long-form articles (kind 30023)
  - /replies - Replies to owner's content
  - /mentions - Posts mentioning the owner
  - /archive - Time-based archives (by year/month)
  - /about - Owner profile (kind 0)
  - /diagnostics - System status and statistics
- Per-section views: gophermap with item type '0' (text) or '1' (submenu) for each event.
- Text rendering: converts Nostr event content to wrapped plain text; shows metadata (author, timestamp, reactions/zaps).
- Thread navigation: parent/replies linked via selectors; indented display.
- Event detail: /event/<id> shows full event with thread context.
- Archives: gophermap by year/month with links to individual posts.
- Diagnostics: text file showing relay status, sync cursors, author counts.

Note: "inbox" and "outbox" are internal concepts for data organization, not exposed as paths.

Gemini Server (gemini://)
- Serves on port 1965 with TLS (self-signed or custom cert).
- Homepage (gemini://host/ or gemini://host): fully configurable via layout.pages.home or layout.sections
  - Default: Auto-generated gemtext with links to all sections
  - Customizable: Can show composed sections, single section, or custom content
- Default sections (same as Gopher, all configurable via layout.sections):
  - /notes - Owner's notes (kind 1, non-replies)
  - /articles - Owner's long-form articles (kind 30023)
  - /replies - Replies to owner's content
  - /mentions - Posts mentioning the owner
  - /archive - Time-based archives (by year/month)
  - /about - Owner profile (kind 0)
  - /diagnostics - System status and statistics
- Per-section views: gemtext document with event links (=> /event/<id>).
- Event rendering: gemtext formatting with headings, quotes, preformatted blocks; reactions/zaps shown as text.
- Event detail: gemini://host/event/<id> shows full event with thread context.
- Thread navigation: links to parent and child replies.
- Input support: search queries, filter selection via Gemini input (status 10).
- Archives: gemtext index by year/month with links.
- Diagnostics: gemtext page with relay/sync status.

Note: "inbox" and "outbox" are internal concepts for data organization, not exposed as paths.

Finger Server (RFC 742)
- Serves on port 79; responds to finger queries.
- Query format: "npub@host" or "username@host" (maps to followed users).
- Response: plain text with owner profile (from kind 0), .plan field (about/bio), recent notes (last 5 kind 1 events).
- Limited to owner + top N followed users (configured via protocols.finger.max_users).
- Shows interaction counts (followers, following, recent zaps/reactions if available).

Content Transformation
- Nostr events → Gopher text: convert markdown to plain text with configurable formatting (see markdown_conversion.md).
  - Headings: underline or uppercase style
  - Bold: UPPERCASE or **preserve asterisks**
  - Links: "text <url>" or "text (url)" format
  - Code blocks: indent or wrap with separators
  - Line wrapping at 70 chars (configurable)
- Nostr events → Gemtext: convert markdown to gemtext format (see markdown_conversion.md).
  - Map headings: # ## ### (flatten deeper levels)
  - Extract inline links to separate => lines
  - Convert ordered lists to unordered (* 1. item)
  - Preserve code blocks and quotes
  - Optional line wrapping at 80 chars
- Nostr events → Finger response: strip all markdown, compact format with timestamps and summaries.
  - Remove all formatting syntax
  - Truncate to max length (default 500 chars)
  - Preserve bare URLs optionally

See markdown_conversion.md for detailed conversion rules and configuration options.

Caching
- Cache rendered gophermaps, gemtext pages, and finger responses per TTL.
- Invalidate on new events, profile updates, or interaction changes.
- Serve stale content if sync is temporarily unavailable.
