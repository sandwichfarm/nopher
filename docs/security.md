# Security Guide

This document describes the security features, best practices, and configuration for nophr.

## Table of Contents

- [Security Architecture](#security-architecture)
- [Deny List](#deny-list)
- [Rate Limiting](#rate-limiting)
- [Input Validation](#input-validation)
- [Secret Management](#secret-management)
- [Content Filtering](#content-filtering)
- [Best Practices](#best-practices)
- [Security Checklist](#security-checklist)

## Security Architecture

nophr implements defense-in-depth security with multiple layers:

1. **Input Validation** - All user input is validated before processing
2. **Rate Limiting** - Prevents abuse and DoS attacks
3. **Deny Lists** - Blocks unwanted pubkeys and content
4. **Content Filtering** - Filters out banned words and patterns
5. **Secret Management** - Secure handling of private keys (env-only, never disk)
6. **Sanitization** - Removes dangerous characters from all inputs

## Deny List

### Overview

The deny list allows you to block specific Nostr pubkeys from appearing in your gateway.

### Configuration

```yaml
security:
  denylist:
    enabled: true
    pubkeys:
      - "deadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678"
      - "cafebabe1234567890abcdef1234567890abcdef1234567890abcdef12345678"
```

### Usage

```go
// Create deny list
dl := security.NewDenyList([]string{
    "pubkey1",
    "pubkey2",
})

// Check if pubkey is denied
if dl.IsPubkeyDenied(pubkey) {
    // Handle denied pubkey
}

// Filter events
allowedEvents := dl.FilterEvents(events)

// Dynamically add/remove pubkeys
dl.AddPubkey("new_blocked_pubkey")
dl.RemovePubkey("unblocked_pubkey")
```

### Thread Safety

All deny list operations are thread-safe using RWMutex for concurrent reads.

## Rate Limiting

### Overview

Rate limiting prevents abuse by limiting the number of requests per client per time window.

Uses a token bucket algorithm with automatic refill.

### Configuration

```yaml
security:
  ratelimit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
```

### Per-Protocol Rate Limits

Different protocols can have different rate limits:

```yaml
security:
  ratelimit:
    gopher:
      requests_per_minute: 30
      burst_size: 5
    gemini:
      requests_per_minute: 60
      burst_size: 10
    finger:
      requests_per_minute: 20
      burst_size: 3
```

### Usage

```go
// Create rate limiter
rl := security.NewRateLimiter(60, time.Minute)
defer rl.Close()

// Check if client is allowed
clientID := getClientIP(conn)
if !rl.Allow(clientID) {
    return errors.New("rate limit exceeded")
}

// Get current limit status
remaining, resetTime := rl.GetLimit(clientID)
```

### Multi-Rate Limiter

For managing multiple rate limiters:

```go
mrl := security.NewMultiRateLimiter()

// Add different limiters for different purposes
mrl.AddLimiter("gopher", security.NewRateLimiter(30, time.Minute))
mrl.AddLimiter("gemini", security.NewRateLimiter(60, time.Minute))

// Check specific limiter
if !mrl.Allow("gopher", clientID) {
    return errors.New("gopher rate limit exceeded")
}
```

### Cleanup

Rate limiters automatically clean up old client buckets to prevent memory leaks.
The cleanup interval is 5 minutes by default and removes buckets inactive for 2x the window duration.

## Input Validation

### Overview

All input is validated before processing to prevent injection attacks and other security issues.

### Gopher Selector Validation

Protects against:
- CRLF injection (`\r\n`)
- Null byte injection (`\x00`)
- Directory traversal (`..`)
- Oversized selectors (max 1024 bytes)

```go
validator := security.NewValidator()

if err := validator.ValidateGopherSelector(selector); err != nil {
    return err
}
```

### Gemini Path Validation

Protects against:
- Directory traversal
- Invalid URL encoding
- Oversized paths (max 4096 bytes)

```go
if err := validator.ValidateGeminiPath(path); err != nil {
    return err
}
```

### Pubkey Validation

Validates Nostr pubkeys (hex format):
- Must be exactly 64 characters
- Must be valid hexadecimal

```go
if err := validator.ValidatePubkey(pubkey); err != nil {
    return err
}
```

### Npub Validation

Validates Nostr npubs (bech32 format):
- Must start with "npub1"
- Must be 63-65 characters
- Must use valid bech32 characters

```go
if err := validator.ValidateNpub(npub); err != nil {
    return err
}
```

### Sanitization

Removes dangerous characters from input:

```go
// Sanitize general input
clean := validator.SanitizeInput(userInput)

// Sanitize file paths
cleanPath := validator.SanitizePath(filePath)
```

### Combined Sanitization and Validation

```go
sanitizer := security.NewInputSanitizer()

// Sanitize and validate in one step
selector, err := sanitizer.SanitizeAndValidateSelector(rawSelector)
if err != nil {
    return err
}
```

## Secret Management

### Overview

nophr handles secrets (private keys) securely:
- **Environment variables only** - Never read from config files
- **Memory only** - Never written to disk
- **Automatic redaction** - Secrets are redacted in logs
- **Secure cleanup** - Memory is overwritten when clearing secrets

### Loading Secrets

```bash
# Set environment variable
export NOPHR_NSEC="nsec1..."

# Start nophr (will load from env)
./nophr
```

### Usage

```go
sm := security.NewSecretManager()

// Load from environment
nsec, err := sm.LoadNsecFromEnv()
if err != nil {
    return err
}

// Store in memory
sm.Set("MY_SECRET", "secret_value")

// Retrieve
value, exists := sm.Get("MY_SECRET")

// Clear all secrets (overwrites with zeros first)
sm.Clear()
```

### Redaction

Secrets are automatically redacted in logs:

```go
// Redact a secret for logging
redacted := sm.Redact("my_secret_key_12345")
// Output: "my_s...2345"
```

### SecureString Type

Use SecureString to prevent accidental logging:

```go
ss := security.NewSecureString("my_secret")

// String() returns redacted version
fmt.Println(ss.String()) // "my_s...cret"

// Get() returns actual value
secret := ss.Get() // "my_secret"

// JSON marshaling is automatically redacted
json.Marshal(ss) // "\"my_s...cret\""

// Clear when done
ss.Clear()
```

### Secret Validation

```go
sv := security.NewSecretValidator()

// Validate nsec format
if err := sv.ValidateNsec(nsec); err != nil {
    return err
}

// Check for leaked secrets in logs
leaks := sv.CheckForLeakedSecrets(logMessage)
if len(leaks) > 0 {
    // Warning: potential secret leak
}
```

### Safe Logging

```go
sl := security.NewSafeLogger()

// Sanitize log messages
safe := sl.SanitizeMessage(message)

// Check for secrets before logging
if err := sl.CheckMessage(message); err != nil {
    // Don't log this message
}
```

## Content Filtering

### Overview

Filter out unwanted content based on banned words or patterns.

### Configuration

```yaml
security:
  content_filter:
    enabled: true
    banned_words:
      - "spam"
      - "scam"
      - "malware"
```

### Usage

```go
cf := security.NewContentFilter([]string{"spam", "scam"})

// Check content
if cf.ContainsBannedContent(text) {
    // Handle banned content
}

// Filter events
if cf.IsEventFiltered(event) {
    // Event contains banned content
}

// Add words dynamically
cf.AddBannedWord("new_banned_word")
```

### Combined Filtering

Use both deny list and content filtering:

```go
dl := security.NewDenyList(blockedPubkeys)
cf := security.NewContentFilter(bannedWords)
combined := security.NewCombinedFilter(dl, cf)

// Check if event is allowed
if !combined.IsEventAllowed(event) {
    // Event is blocked
}

// Filter event list
allowed := combined.FilterEvents(events)
```

### Security Policy

Define a complete security policy:

```go
policy := &security.SecurityPolicy{
    DenyListPubkeys: []string{"pubkey1", "pubkey2"},
    BannedWords:     []string{"spam", "scam"},
    AllowAnonymous:  false,
    RequireNIP05:    true,
}

enforcer := security.NewEnforcer(policy)

// Enforce on single event
if err := enforcer.EnforceEvent(ctx, event); err != nil {
    // Event denied
}

// Filter event list
allowed := enforcer.EnforceEvents(ctx, events)
```

## Best Practices

### 1. Always Use HTTPS/TLS

For Gemini protocol, TLS is required. For Gopher and Finger, use a reverse proxy:

```nginx
server {
    listen 443 ssl;
    server_name gopher.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:7000;
    }
}
```

### 2. Run as Non-Root User

Never run nophr as root. Use systemd with `User=nophr`:

```ini
[Service]
User=nophr
Group=nophr
```

### 3. Enable Privilege Separation

Use systemd security features:

```ini
[Service]
# Filesystem protection
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/nophr

# Network isolation
PrivateNetwork=false
RestrictAddressFamilies=AF_INET AF_INET6

# Privilege restrictions
NoNewPrivileges=true
PrivateTmp=true
```

### 4. Use Environment Variables for Secrets

Never put secrets in config files:

```bash
# Good
export NOPHR_NSEC="nsec1..."

# Bad - never do this
echo "nsec: nsec1..." >> config.yaml
```

### 5. Enable Rate Limiting

Always enable rate limiting in production:

```yaml
security:
  ratelimit:
    enabled: true
    requests_per_minute: 60
```

### 6. Regularly Update Deny List

Monitor for spam/abuse and update your deny list:

```bash
# Add to deny list
curl -X POST http://localhost:8080/admin/denylist \
  -d '{"pubkey": "deadbeef..."}'
```

### 7. Monitor Logs for Security Events

Check logs regularly for:
- Rate limit violations
- Invalid input attempts
- Denied pubkey access attempts

### 8. Keep Dependencies Updated

```bash
go get -u ./...
go mod tidy
```

### 9. Use Strong TLS Configuration

For Gemini TLS:

```go
tlsConfig := &tls.Config{
    MinVersion:               tls.VersionTLS13,
    CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
    PreferServerCipherSuites: true,
}
```

### 10. Implement Request Timeouts

```yaml
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
```

## Security Checklist

Before deploying to production:

- [ ] Secrets are loaded from environment variables only
- [ ] Running as non-root user
- [ ] Rate limiting is enabled
- [ ] Input validation is enabled
- [ ] Deny list is configured
- [ ] Content filtering is configured
- [ ] HTTPS/TLS is enabled (for public access)
- [ ] Systemd security features are enabled
- [ ] Request timeouts are configured
- [ ] Logs are being monitored
- [ ] Regular backups are configured
- [ ] Dependencies are up to date
- [ ] Firewall rules are configured
- [ ] Reverse proxy is configured (if needed)
- [ ] Error messages don't leak sensitive info

## Reporting Security Issues

If you discover a security vulnerability, please:

1. **Do NOT** open a public issue
2. Email security@example.com with details
3. Include steps to reproduce
4. Allow time for a fix before public disclosure

## Security Updates

Check for security updates regularly:

- Subscribe to the GitHub repository releases
- Monitor the security mailing list
- Follow @nophr on Nostr for announcements

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Nostr NIPs](https://github.com/nostr-protocol/nips)
- [Go Security Policy](https://go.dev/security/policy)
- [Systemd Security](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#Security)
