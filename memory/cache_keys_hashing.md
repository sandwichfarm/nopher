Cache Keys and Config Hashing

Overview
- All cache keys live under a namespaced prefix computed from a stable config hash.
- Keys are deterministic, short, and collision-resistant; they avoid secrets and machine-local paths.
- When relevant config/theme/layout changes, the namespace changes automatically, making old caches cold.

Config Hash (cfg)
Inputs
- App version: semantic or git sha (e.g., 0.1.0 or abc1234) to invalidate on rendering/data model changes.
- Effective config (normalized): site, discovery, sync, caching, layout, export, storage driver type; exclude secrets.
- Theme config (not assets): theme name and selected options.

Normalization
1) Start from the loaded config object.
2) Remove secret-bearing keys: identity.nsec, storage.postgres_url.
3) Remove environment-only and transient fields (diagnostic toggles, file paths if not content-affecting).
4) Expand macros to canonical forms (e.g., authors:[owner] -> authors:[<owner npub>]).
5) Sort keys recursively; sort array values; drop null/false/defaults when safe.
6) Serialize to compact JSON without whitespace.

Hash
- cfg = hex(SHA-256("ver="+app_version+"\n"+normalized_json))[0:12]
- Length (12) is configurable; use 16 for very large deployments.
- Optionally include a salt via NOPHR_CACHE_SALT to segregate shared Redis across multiple tenants.

Theme and Layout Hashes (sub-hashes)
- theme_h = hex(SHA-256(theme_name + JSON(theme_options)))[0:8]
- layout_h = hex(SHA-256(JSON(layout.sections) + JSON(layout.pages)))[0:8]
- Use these when keys depend on rendering templates or layout composition.

Query Hash (q)
Purpose
- Compact, stable identity for a section query (filters + transforms + sort + paging window).

Normalization
1) Build a query descriptor object with keys: source, filter, transform, sort, limit, page, page_size.
2) Expand macros (e.g., owner -> npub string); normalize authors lists; resolve section:<id> references.
3) Canonicalize: sort keys; sort arrays; remove defaults (e.g., group_by_thread:false) and nulls.
4) Stable stringify to compact JSON.

Hash
- q = hex(SHA-256(JSON))[0:12]

Key Format (with cfg namespace)
- v:{cfg}:render:event:{event_id}:t:{template_h}
- v:{cfg}:render:param:{kind}:{pubkey}:d:{d_tag}:t:{template_h}
- v:{cfg}:section:{section_id}:page:{page}:q:{q}
- v:{cfg}:feed:{section_id}:{format}:q:{q}
- v:{cfg}:page:{page_id}:layout:{layout_h}
- v:{cfg}:agg:{event_id}
- v:{cfg}:etag:{resource_id}

Values (examples)
- render:* -> HTML bytes + content metadata (lang, created_at)
- section:* -> { ids: [event_id...], created_at_max, count }
- feed:* -> bytes + etag + last_modified
- agg:* -> { reply_count, reaction_total, reaction_counts_json, zap_sats_total, last_interaction_at, ver }

ETag Derivation
Goals
- Cheap to compute, stable for identical visible results, sensitive to aggregates for visible items.

Section/Feed ETag
1) Take ordered list of event ids (top N items on the page or feed).
2) Build a small vector per id: [reply_count, reaction_total, zap_sats_total, last_interaction_at].
3) Compute base = SHA-1(ids concatenated with commas).
4) Compute overlay = SHA-1(JSON of vectors with ids order preserved).
5) Compute lm = max(created_at) across items.
6) etag = base[0:8] + '-' + overlay[0:8] + '-' + to_hex(lm)
- Store alongside cached section/feed; send as HTTP ETag.

Event Render Version (template_h)
- template_h = hex(SHA-256(theme_h + template_name + renderer_version + markdown_opts))[0:8]
- Bump renderer_version when changing markdown/HTML rendering.

Collisions and Safety
- 12-hex (~48 bits) is sufficient in practice; use 16-hex in shared Redis with many tenants.
- Keys always carry cfg to avoid cross-config bleed.
- If a collision is ever detected (value shape mismatch), fall back to recompute and optionally increase key length.

Pseudocode (TypeScript-like)

function cfgHash(appVersion, cfg) {
  const norm = normalizeConfig(cfg);             // canonical object per steps above
  const payload = `ver=${appVersion}\n` + JSON.stringify(norm);
  return sha256hex(payload).slice(0, 12);
}

function queryHash(desc) {
  const norm = normalizeQuery(desc);             // expand macros, sort keys/arrays, drop defaults
  return sha256hex(JSON.stringify(norm)).slice(0, 12);
}

function sectionKey(cfgH, sectionId, page, q) {
  return `v:${cfgH}:section:${sectionId}:page:${page}:q:${q}`;
}

function eventRenderKey(cfgH, eventId, templateH) {
  return `v:${cfgH}:render:event:${eventId}:t:${templateH}`;
}

function etagForSection(ids, aggVecs, maxCreatedAt) {
  const base = sha1hex(ids.join(','));
  const overlay = sha1hex(JSON.stringify(aggVecs));
  return `${base.slice(0,8)}-${overlay.slice(0,8)}-${maxCreatedAt.toString(16)}`;
}

