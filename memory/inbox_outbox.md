Inbox and Outbox

Outbox
- Content authored by the owner.
- Includes: notes (1), long-form articles (30023), reposts (6), reactions (7, optional), zaps (9735, optional).
- Optional publisher can sign and publish to write relays discovered via NIP-65.

Inbox
- Content targeting the owner or the owner's content.
- Replies: NIP-10 events referencing owner's posts via #e root/reply, or mentioning owner via #p.
- Reactions: kind 7 referencing owner's events; aggregate counts and latest reactors.
- Zaps: kind 9735; parse bolt11 invoice amounts when present; attribute to target event or profile.

Aggregation and Display
- Group by thread (root) for replies and interactions.
- Collapse reposts to avoid noise.
- Noise filters: min zap sats; allowed reaction chars; deny-listed pubkeys.
- Diagnostics: show when items are omitted due to filters or caps.
