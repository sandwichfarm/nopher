# Nopher Testing Instructions

## Prerequisites

- Go 1.21 or later
- SQLite3 command-line tool
- netcat (`nc`) for Gopher protocol testing
- OpenSSL for Gemini protocol testing (optional)
- Valid Nostr npub in `test-config.yaml`

## Build and Start

```bash
# Build the binary
go build -o nopher ./cmd/nopher

# Create test data directory if needed
mkdir -p test-data

# Start nopher with test configuration
./nopher --config test-config.yaml

# Or run in background with logging
./nopher --config test-config.yaml > nopher.log 2>&1 &
```

## Wait for Initial Sync

Nopher needs 20-30 seconds to sync events from relays on first start.

```bash
# Watch the logs to see sync progress
tail -f nopher.log

# You should see messages like:
# [NOSTR CLIENT] Received 100 events so far...
# [NOSTR CLIENT] Received 200 events so far...
# etc.

# Check database population
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM event;"
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM aggregates WHERE reaction_total > 0;"
```

Expected results after sync:
- 500+ events in database
- 300+ aggregates with reactions
- 90+ kind 1 notes

## Gopher Protocol Tests (Port 7070)

### Test 1: Home Page
```bash
printf "/\r\n" | nc localhost 7070
```

**Expected output:**
- Title: "Nopher - Nostr Gateway"
- Menu items: Notes, Articles, Replies, Mentions, Diagnostics
- Gopher directory entries (type `1`)

### Test 2: Notes Endpoint
```bash
printf "/notes\r\n" | nc localhost 7070 | head -30
```

**Expected output:**
- Title: "Notes"
- List of owner's notes (kind 1, non-replies only)
- Each note shows:
  - Number and first line of content
  - Author pubkey (truncated)
  - Timestamp (relative or absolute)
  - Interactions if present
- Links to individual notes (`0` type for text files)

### Test 3: Articles Endpoint
```bash
printf "/articles\r\n" | nc localhost 7070
```

**Expected output:**
- Title: "Articles"
- "No notes yet" (if no kind 30023 articles exist)
- Or list of articles with same format as notes

### Test 4: Replies Endpoint
```bash
printf "/replies\r\n" | nc localhost 7070 | head -30
```

**Expected output:**
- Title: "Replies"
- List of replies to owner's content
- Each reply has:
  - Content preview
  - Author and timestamp
  - Possible interactions

### Test 5: Mentions Endpoint
```bash
printf "/mentions\r\n" | nc localhost 7070 | head -30
```

**Expected output:**
- Title: "Mentions"
- List of all mentions of owner (includes both replies and non-reply mentions)
- Similar format to replies list

### Test 6: Legacy /outbox Redirect
```bash
printf "/outbox\r\n" | nc localhost 7070 | head -15
```

**Expected output:**
- Should redirect to Notes
- Same output as `/notes` endpoint

### Test 7: Legacy /inbox Redirect
```bash
printf "/inbox\r\n" | nc localhost 7070 | head -15
```

**Expected output:**
- Should redirect to Replies
- Same output as `/replies` endpoint

### Test 8: Individual Note View
```bash
# Get a note ID from the notes list first
NOTE_ID=$(printf "/notes\r\n" | nc localhost 7070 | grep "^0" | head -1 | awk '{print $2}' | sed 's|/note/||')

# View the note
printf "/note/$NOTE_ID\r\n" | nc localhost 7070
```

**Expected output:**
- Note header with author and timestamp
- Full note content
- Separator line (70 dashes)
- Interactions section (if note has interactions)
- Links to view thread and home

### Test 9: Thread View
```bash
# Use same note ID from above
printf "/thread/$NOTE_ID\r\n" | nc localhost 7070
```

**Expected output:**
- Thread header
- Root post with full content
- Replies section listing all replies
- Each reply indented with "↳" symbol

### Test 10: Diagnostics Page
```bash
printf "/diagnostics\r\n" | nc localhost 7070
```

**Expected output:**
- Server status: Running
- Host and port information
- Storage status: Connected

## Enhanced Aggregate Rendering Tests

### Test 11: Note with Reactions Breakdown

First, add a test aggregate to see the enhanced rendering:

```bash
# Get first note ID
NOTE_ID=$(sqlite3 test-data/nopher.db "SELECT id FROM event WHERE kind = 1 AND pubkey = (SELECT DISTINCT pubkey FROM event WHERE kind = 1 LIMIT 1) LIMIT 1;")

# Add test aggregate with multiple reaction types
sqlite3 test-data/nopher.db "INSERT OR REPLACE INTO aggregates (event_id, reply_count, reaction_total, reaction_counts_json, zap_sats_total, last_interaction_at) VALUES ('$NOTE_ID', 3, 7, '{\"❤️\":3,\"🔥\":2,\"🚀\":1,\"👍\":1}', 21000, $(date +%s));"

# View the note
printf "/note/$NOTE_ID\r\n" | nc localhost 7070
```

**Expected output:**
```
Interactions: 3 replies, 7 reactions (❤️ 3, 🔥 2, 🚀 1, 👍 1), 21.0K sats zapped
```

The reaction breakdown should show:
- Total reaction count: `7 reactions`
- Breakdown in parentheses: `(❤️ 3, 🔥 2, 🚀 1, 👍 1)`
- Reply count and zap amount

### Test 12: Reaction Breakdown in List View
```bash
printf "/notes\r\n" | nc localhost 7070 | head -10
```

**Expected output:**
First note should show the enhanced interactions line with reaction breakdown.

## Gemini Protocol Tests (Port 1965) [Optional]

### Test 1: Gemini Home Page
```bash
echo "gemini://localhost/" | openssl s_client -connect localhost:1965 -quiet -ign_eof 2>/dev/null
```

**Expected output:**
- Status: `20 text/gemini`
- Gemtext formatted home page with links

### Test 2: Gemini Notes Endpoint
```bash
echo "gemini://localhost/notes" | openssl s_client -connect localhost:1965 -quiet -ign_eof 2>/dev/null | head -30
```

**Expected output:**
- Gemtext formatted notes list
- Links using `=>` syntax

## Database Verification Tests

### Test 13: Verify Event Storage
```bash
# Count total events
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM event;"

# Count by kind
sqlite3 test-data/nopher.db "SELECT kind, COUNT(*) FROM event GROUP BY kind ORDER BY COUNT(*) DESC;"

# Show sample events
sqlite3 test-data/nopher.db "SELECT id, kind, substr(content, 1, 50) FROM event WHERE kind = 1 LIMIT 5;"
```

**Expected results:**
- Multiple event kinds (1, 7, 0, 3, etc.)
- Kind 1 (notes) should have significant count
- Kind 7 (reactions) should have high count

### Test 14: Verify Aggregate Calculation
```bash
# Count aggregates
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM aggregates;"

# Show aggregates with reactions
sqlite3 test-data/nopher.db "SELECT event_id, reply_count, reaction_total, reaction_counts_json FROM aggregates WHERE reaction_total > 0 LIMIT 5;"

# Find events with multiple reaction types
sqlite3 test-data/nopher.db "SELECT event_id, reaction_total, reaction_counts_json FROM aggregates WHERE reaction_counts_json LIKE '%,%' LIMIT 5;"
```

**Expected results:**
- Many aggregates with reaction data
- `reaction_counts_json` contains emoji breakdown like `{"❤️":1,"🔥":2}`

### Test 15: Verify Owner's Content
```bash
# Get owner's pubkey from config
OWNER_PUBKEY=$(sqlite3 test-data/nopher.db "SELECT DISTINCT pubkey FROM event WHERE kind = 1 LIMIT 1;")

# Count owner's notes
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM event WHERE kind = 1 AND pubkey = '$OWNER_PUBKEY';"

# Count owner's replies
sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM event WHERE kind = 1 AND pubkey = '$OWNER_PUBKEY' AND json_extract(tags, '$[0][0]') = 'e';"
```

**Expected results:**
- Owner should have multiple notes
- Some notes should be replies (have 'e' tags)

## Performance Tests

### Test 16: Query Response Time
```bash
# Time a notes query
time printf "/notes\r\n" | nc localhost 7070 > /dev/null
```

**Expected result:**
- Should complete in under 1 second

### Test 17: Concurrent Requests
```bash
# Send multiple requests in parallel
for i in {1..10}; do
  printf "/notes\r\n" | nc localhost 7070 > /dev/null &
done
wait
```

**Expected result:**
- All requests should complete successfully
- No errors in logs

## Error Handling Tests

### Test 18: Invalid Note ID
```bash
printf "/note/invalid_id_12345\r\n" | nc localhost 7070
```

**Expected output:**
- Error message: "Note not found: invalid_id_12345"
- Link back to home

### Test 19: Unknown Route
```bash
printf "/unknown_route\r\n" | nc localhost 7070
```

**Expected output:**
- Error message: "Unknown selector: unknown_route"
- Link back to home

## Cleanup

```bash
# Stop nopher
pkill -f 'nopher --config'

# Optional: Remove test data
# rm -rf test-data/
```

## Troubleshooting

### No events syncing
- Check logs: `tail -f nopher.log`
- Verify relay connectivity: Look for relay connection messages
- Check seed relays in `test-config.yaml`
- Wait longer (30-60 seconds) for initial sync

### Empty query results
- Verify database has events: `sqlite3 test-data/nopher.db "SELECT COUNT(*) FROM event;"`
- Check owner pubkey matches config npub (hex vs bech32)
- Look for errors in logs

### Connection refused
- Verify nopher is running: `ps aux | grep nopher`
- Check ports 7070 (Gopher) and 1965 (Gemini) are not in use
- Review startup logs for binding errors

### Duplicate event errors
- Normal during sync when same events come from multiple relays
- Not a problem - duplicates are safely rejected

## Success Criteria

All tests pass if:
- ✅ Home page displays with menu
- ✅ `/notes` shows owner's notes (non-replies)
- ✅ `/replies` shows replies to owner
- ✅ `/mentions` shows mentions of owner
- ✅ `/articles` shows articles or "No notes yet"
- ✅ Legacy `/inbox` and `/outbox` redirect properly
- ✅ Individual note view shows full content
- ✅ Thread view shows replies
- ✅ Reaction breakdown displays as: `7 reactions (❤️ 3, 🔥 2, 🚀 1, 👍 1)`
- ✅ Database contains events and aggregates
- ✅ No critical errors in logs
