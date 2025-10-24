Distribution Strategy

Goal: Make Nopher as easy to install as possible across different platforms and use cases.

Single Binary Distribution

Yes! The entire application can be packaged as a single static binary with embedded resources.

Go Embed Feature (Go 1.16+)
- Use //go:embed directive to embed default configuration templates
- Embed example gophermap templates, gemtext templates
- Embed TLS certificate generation utilities
- No external file dependencies required (except user config and database)

Benefits:
- Zero dependencies: just download and run
- Cross-platform: compile for Linux, macOS, Windows, BSD
- Simple deployment: scp binary to server, chmod +x, run
- Easy updates: replace binary, restart service
- Perfect for gopherholes: minimalist distribution for minimalist protocols

What Goes in the Binary:
- Application code (sync engine, protocol servers, renderers)
- Embedded Khatru and eventstore libraries
- Default configuration template (nopher.example.yaml)
- Example layouts and section definitions
- Markdown conversion libraries
- TLS certificate generator for Gemini

What Stays External:
- User configuration file (nopher.yaml)
- Database files (SQLite or LMDB)
- TLS certificates (if user-provided)
- Logs

Distribution Channels

1. GitHub Releases (Primary)
- Pre-built binaries for major platforms via GoReleaser
- Platforms:
  - Linux: amd64, arm64, armv7
  - macOS: amd64 (Intel), arm64 (Apple Silicon)
  - FreeBSD: amd64, arm64
  - OpenBSD: amd64
  - Windows: amd64 (for testing/development)
- Archive formats: .tar.gz (Unix), .zip (Windows)
- Include: binary, example config, README, LICENSE
- Checksums (SHA256) for verification
- GPG signatures for security

2. Package Managers

Homebrew (macOS and Linux)
```bash
brew tap sandwich/nopher
brew install nopher
```
- GoReleaser auto-generates Homebrew formula
- Updates handled via brew upgrade
- Easy uninstall: brew uninstall nopher

APT (Debian/Ubuntu)
```bash
echo 'deb [trusted=yes] https://apt.nopher.io/ /' | sudo tee /etc/apt/sources.list.d/nopher.list
sudo apt update
sudo apt install nopher
```
- GoReleaser creates .deb packages
- Includes systemd service file
- Post-install script creates /etc/nopher/nopher.yaml template

RPM (Fedora/RHEL/Rocky/Alma)
```bash
sudo dnf config-manager --add-repo https://rpm.nopher.io/nopher.repo
sudo dnf install nopher
```
- GoReleaser creates .rpm packages
- Includes systemd service file

AUR (Arch Linux)
```bash
yay -S nopher-bin  # binary package
# or
yay -S nopher      # build from source
```
- Maintain PKGBUILD in AUR
- Community-maintained option

Snap (Ubuntu/Linux)
```bash
sudo snap install nopher
```
- Sandboxed environment
- Auto-updates by default
- May require connection permissions for network protocols

3. Docker / Container Images

Docker Hub
```bash
docker pull sandwich/nopher:latest
docker pull sandwich/nopher:v1.0.0
docker pull sandwich/nopher:v1-alpine
```

GitHub Container Registry
```bash
docker pull ghcr.io/sandwich/nopher:latest
```

Tags:
- latest: latest stable release
- v1.0.0: specific version
- v1-alpine: Alpine Linux base (smaller)
- edge: latest commit on main branch

Docker Compose Example:
```yaml
version: '3.8'
services:
  nopher:
    image: sandwich/nopher:latest
    ports:
      - "70:70"      # Gopher
      - "1965:1965"  # Gemini
      - "79:79"      # Finger
    volumes:
      - ./nopher.yaml:/etc/nopher/nopher.yaml:ro
      - ./data:/var/lib/nopher
      - ./certs:/etc/nopher/certs:ro
    environment:
      - NOPHER_NSEC=${NOPHER_NSEC}
    restart: unless-stopped
```

Multi-architecture images:
- linux/amd64
- linux/arm64
- linux/arm/v7

4. Building from Source

Go Toolchain Required:
```bash
# Clone repository
git clone https://github.com/sandwich/nopher.git
cd nopher

# Build
go build -o nopher cmd/nopher/main.go

# Or use Makefile
make build

# Install to $GOPATH/bin
make install
```

Build Tags (optional features):
```bash
# Minimal build (SQLite only)
go build -tags sqlite

# LMDB support
go build -tags "sqlite lmdb"

# All features
go build -tags "sqlite lmdb debug"
```

5. One-Line Installers

Bash Script (Linux/macOS):
```bash
curl -fsSL https://get.nopher.io | sh
```

Script behavior:
- Detects OS and architecture
- Downloads appropriate binary from GitHub releases
- Verifies checksum
- Installs to /usr/local/bin (or ~/bin if no sudo)
- Creates example config at ~/.config/nopher/nopher.yaml
- Prints next steps

PowerShell Script (Windows):
```powershell
iwr -useb https://get.nopher.io/install.ps1 | iex
```

6. System Package Details

All packages include:
- Binary: /usr/bin/nopher
- Config: /etc/nopher/nopher.example.yaml
- Data dir: /var/lib/nopher (created, owned by nopher user)
- Systemd service: /etc/systemd/system/nopher.service
- Man page: /usr/share/man/man1/nopher.1.gz
- Documentation: /usr/share/doc/nopher/

Systemd Service File:
```ini
[Unit]
Description=Nopher - Nostr to Gopher/Gemini/Finger Gateway
After=network.target
Documentation=https://github.com/sandwich/nopher

[Service]
Type=simple
User=nopher
Group=nopher
WorkingDirectory=/var/lib/nopher
ExecStart=/usr/bin/nopher --config /etc/nopher/nopher.yaml
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/nopher
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

Reverse Proxy Examples

Caddy (Recommended for Gemini TLS passthrough)

/etc/caddy/Caddyfile:
```
# Gopher proxy (if needed)
:70 {
    reverse_proxy localhost:7070
}

# Gemini - direct TLS passthrough (Nopher handles TLS)
# No proxy needed - Nopher listens on 1965 directly

# Optional: HTTPS admin panel
nopher-admin.example.com {
    reverse_proxy localhost:8080
}
```

Nginx (Gopher proxy if running non-privileged)

/etc/nginx/streams.d/gopher.conf:
```nginx
stream {
    upstream gopher_backend {
        server 127.0.0.1:7070;
    }

    server {
        listen 70;
        proxy_pass gopher_backend;
    }
}
```

Note: Nginx doesn't natively support Gopher or Gemini protocols.
Use for TCP proxying only if running Nopher on non-privileged ports.

HAProxy (TCP load balancing for multiple instances)

/etc/haproxy/haproxy.cfg:
```
frontend gopher
    bind *:70
    mode tcp
    default_backend gopher_servers

backend gopher_servers
    mode tcp
    balance leastconn
    server nopher1 127.0.0.1:7070 check
    server nopher2 127.0.0.1:7071 check
```

systemd Socket Activation (Privilege separation)

Instead of running as root, use systemd socket activation:

/etc/systemd/system/nopher.socket:
```ini
[Unit]
Description=Nopher Socket Activation

[Socket]
ListenStream=70
ListenStream=1965
ListenStream=79

[Install]
WantedBy=sockets.target
```

/etc/systemd/system/nopher.service:
```ini
[Unit]
Description=Nopher Service
Requires=nopher.socket

[Service]
Type=simple
User=nopher
ExecStart=/usr/bin/nopher --systemd-socket
StandardInput=socket
```

This allows Nopher to run as unprivileged user while binding to privileged ports.

GoReleaser Configuration

.goreleaser.yaml:
```yaml
before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: nopher
    main: ./cmd/nopher
    binary: nopher
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - freebsd
      - openbsd
    goarch:
      - amd64
      - arm64
      - arm
    goarm:
      - "7"
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    files:
      - README.md
      - LICENSE
      - configs/nopher.example.yaml
      - docs/*

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

brews:
  - name: nopher
    tap:
      owner: sandwich
      name: homebrew-nopher
    homepage: https://github.com/sandwich/nopher
    description: "Nostr to Gopher/Gemini/Finger gateway"
    license: MIT
    install: |
      bin.install "nopher"
      etc.install "configs/nopher.example.yaml" => "nopher.example.yaml"
    test: |
      system "#{bin}/nopher", "--version"

nfpms:
  - id: nopher-packages
    package_name: nopher
    homepage: https://github.com/sandwich/nopher
    maintainer: Your Name <you@example.com>
    description: Nostr to Gopher/Gemini/Finger gateway
    license: MIT
    formats:
      - deb
      - rpm
      - apk
    bindir: /usr/bin
    contents:
      - src: configs/nopher.example.yaml
        dst: /etc/nopher/nopher.example.yaml
        type: config
      - src: scripts/nopher.service
        dst: /etc/systemd/system/nopher.service
        type: config
    scripts:
      postinstall: scripts/postinstall.sh

dockers:
  - image_templates:
      - "sandwich/nopher:{{ .Version }}-amd64"
      - "sandwich/nopher:latest-amd64"
      - "ghcr.io/sandwich/nopher:{{ .Version }}-amd64"
      - "ghcr.io/sandwich/nopher:latest-amd64"
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - "--platform=linux/amd64"
      - "--label=org.opencontainers.image.title={{ .ProjectName }}"
      - "--label=org.opencontainers.image.version={{ .Version }}"

docker_manifests:
  - name_template: sandwich/nopher:{{ .Version }}
    image_templates:
      - sandwich/nopher:{{ .Version }}-amd64
      - sandwich/nopher:{{ .Version }}-arm64
  - name_template: sandwich/nopher:latest
    image_templates:
      - sandwich/nopher:latest-amd64
      - sandwich/nopher:latest-arm64
```

Dockerfile (Multi-stage)

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w" \
    -o nopher cmd/nopher/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 nopher && \
    adduser -D -u 1000 -G nopher nopher

WORKDIR /app

COPY --from=builder /build/nopher /usr/local/bin/nopher
COPY configs/nopher.example.yaml /etc/nopher/nopher.example.yaml

RUN mkdir -p /var/lib/nopher /etc/nopher/certs && \
    chown -R nopher:nopher /var/lib/nopher /etc/nopher

USER nopher

VOLUME ["/var/lib/nopher", "/etc/nopher"]

EXPOSE 70 1965 79

ENTRYPOINT ["/usr/local/bin/nopher"]
CMD ["--config", "/etc/nopher/nopher.yaml"]
```

Documentation for Users

Quick Start Guide (in README.md):

1. Download binary:
   ```bash
   # Linux/macOS one-liner
   curl -fsSL https://get.nopher.io | sh

   # Or download from releases
   wget https://github.com/sandwich/nopher/releases/download/v1.0.0/nopher_1.0.0_linux_amd64.tar.gz
   tar xzf nopher_1.0.0_linux_amd64.tar.gz
   sudo mv nopher /usr/local/bin/
   ```

2. Create config:
   ```bash
   mkdir -p ~/.config/nopher
   nopher init > ~/.config/nopher/nopher.yaml
   # Edit config with your npub and seed relays
   ```

3. Run:
   ```bash
   nopher --config ~/.config/nopher/nopher.yaml
   ```

4. Test:
   ```bash
   # Gopher
   lynx gopher://localhost

   # Gemini
   amfora gemini://localhost

   # Finger
   finger @localhost
   ```

Installation Matrix

| Method | Platforms | Auto-update | Privileges | Complexity |
|--------|-----------|-------------|------------|------------|
| Binary | All | Manual | User-managed | Lowest |
| Homebrew | macOS/Linux | brew upgrade | User | Low |
| APT/RPM | Linux | apt/dnf | Root | Low |
| Snap | Linux | Automatic | Confined | Low |
| Docker | All | Pull new image | User/root | Medium |
| Source | All | Manual | User | High |

Recommended by Use Case:
- Personal gopherhole: Single binary or Homebrew
- VPS deployment: APT/RPM packages with systemd
- Homelab: Docker Compose
- Development: Build from source
- Multi-instance: Docker Swarm or Kubernetes

Summary

✅ Single binary: Yes, with embedded resources via //go:embed
✅ Package managers: Homebrew, APT, RPM, Snap, AUR
✅ Containers: Docker multi-arch images
✅ One-line installer: Bash/PowerShell scripts
✅ Source: Standard Go build process
✅ Reverse proxies: Example configs for Caddy, Nginx, HAProxy
✅ Systemd integration: Socket activation, service hardening

The Go ecosystem makes all of this achievable with GoReleaser automating most of the distribution pipeline.
