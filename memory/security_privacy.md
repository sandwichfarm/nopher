Security and Privacy

Secrets
- Keep nsec and DB credentials in environment variables only.
- Never write secrets to config files or logs.

Publishing
- Outbox publishing is opt-in. Dry-run mode available.
- Respect discovered write relays from NIP-65; allow operator override.

Data Handling
- Deny-list abusive pubkeys and relays.
- Limit FOAF expansion and thread depth to avoid data explosion.
- Display zap amounts from receipts; do not poll wallets unless explicitly configured.

Privacy
- Do not leak internal state in public pages (e.g., exact relay list) unless enabled.
- Redact or hash identifiers in diagnostics if diagnostics are public.
