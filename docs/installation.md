# Nopher Installation Guide

This guide covers different ways to install and deploy Nopher.

## Table of Contents

- [Quick Install](#quick-install)
- [Package Managers](#package-managers)
- [Docker](#docker)
- [From Source](#from-source)
- [Systemd Service](#systemd-service)
- [Reverse Proxy Setup](#reverse-proxy-setup)
- [Configuration](#configuration)

## Quick Install

The easiest way to install Nopher is using our one-line installer:

```bash
curl -sSL https://nopher.io/install.sh | sh
```

This will:
- Detect your platform and architecture
- Download the latest release
- Install to `/usr/local/bin/nopher`
- Create example configuration at `~/.config/nopher/nopher.yaml`

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew tap sandwichfarm/tap
brew install nopher
```

### Debian/Ubuntu (DEB)

```bash
# Download the latest .deb package
curl -LO https://github.com/sandwichfarm/nopher/releases/download/v0.1.0/nopher_0.1.0_amd64.deb

# Install
sudo dpkg -i nopher_0.1.0_amd64.deb

# The post-install script will:
# - Create 'nopher' system user
# - Set up directories (/var/lib/nopher, /etc/nopher)
# - Install systemd service
```

### RHEL/CentOS/Fedora (RPM)

```bash
# Download the latest .rpm package
curl -LO https://github.com/sandwichfarm/nopher/releases/download/v0.1.0/nopher_0.1.0_amd64.rpm

# Install
sudo yum localinstall nopher_0.1.0_amd64.rpm
# or
sudo dnf install nopher_0.1.0_amd64.rpm
```

### Alpine (APK)

```bash
# Download the latest .apk package
curl -LO https://github.com/sandwichfarm/nopher/releases/download/v0.1.0/nopher_0.1.0_amd64.apk

# Install
sudo apk add --allow-untrusted nopher_0.1.0_amd64.apk
```

## Docker

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/sandwichfarm/nopher.git
cd nopher

# Copy example config
cp configs/nopher.example.yaml configs/nopher.yaml

# Edit your configuration
nano configs/nopher.yaml

# Set your NSEC (never commit this!)
export NOPHER_NSEC="nsec1..."

# Start the services
docker-compose up -d

# View logs
docker-compose logs -f nopher

# Stop the services
docker-compose down
```

### Using Docker directly

```bash
# Pull the latest image
docker pull ghcr.io/sandwichfarm/nopher:latest

# Run with config
docker run -d \
  --name nopher \
  -p 70:70 \
  -p 1965:1965 \
  -p 79:79 \
  -v ./nopher.yaml:/etc/nopher/nopher.yaml:ro \
  -v nopher-data:/var/lib/nopher \
  -e NOPHER_NSEC="nsec1..." \
  ghcr.io/sandwichfarm/nopher:latest
```

### Building from source with Docker

```bash
git clone https://github.com/sandwichfarm/nopher.git
cd nopher

# Build the image
docker build -t nopher:latest .

# Run
docker run -d \
  --name nopher \
  -p 70:70 \
  -p 1965:1965 \
  -p 79:79 \
  -v ./nopher.yaml:/etc/nopher/nopher.yaml:ro \
  -e NOPHER_NSEC="nsec1..." \
  nopher:latest
```

## From Source

### Prerequisites

- Go 1.21 or later
- Git

### Build and Install

```bash
# Clone the repository
git clone https://github.com/sandwichfarm/nopher.git
cd nopher

# Build
go build -o nopher ./cmd/nopher

# Install to /usr/local/bin
sudo mv nopher /usr/local/bin/

# Or keep it local
./nopher --help
```

### Development Mode

```bash
# Run directly with Go
go run ./cmd/nopher --config ./test-config.yaml

# Build for development
go build -o nopher ./cmd/nopher
./nopher --config ./test-config.yaml
```

## Systemd Service

After installing via package manager, enable and start the service:

```bash
# Create your config
sudo cp /etc/nopher/nopher.example.yaml /etc/nopher/nopher.yaml
sudo nano /etc/nopher/nopher.yaml

# Set your NSEC (secure method)
echo 'NOPHER_NSEC="nsec1..."' | sudo tee /etc/default/nopher

# Enable and start service
sudo systemctl enable nopher
sudo systemctl start nopher

# Check status
sudo systemctl status nopher

# View logs
sudo journalctl -u nopher -f
```

### Manual Systemd Setup

If you installed from source or via the installer script:

```bash
# Copy the service file
sudo cp scripts/systemd/nopher.service /etc/systemd/system/

# Create nopher user
sudo useradd --system --no-create-home --shell /bin/false nopher

# Create directories
sudo mkdir -p /var/lib/nopher /etc/nopher
sudo chown nopher:nopher /var/lib/nopher

# Copy config
sudo cp configs/nopher.example.yaml /etc/nopher/nopher.yaml

# Edit config
sudo nano /etc/nopher/nopher.yaml

# Reload systemd
sudo systemctl daemon-reload

# Enable and start
sudo systemctl enable nopher
sudo systemctl start nopher
```

## Reverse Proxy Setup

For production deployments, it's recommended to use a reverse proxy for TLS termination.

### Caddy (Automatic HTTPS)

Caddy automatically handles TLS certificates via Let's Encrypt.

```bash
# Install Caddy
sudo apt install caddy  # Debian/Ubuntu
# or
brew install caddy      # macOS

# Copy the example Caddyfile
sudo cp examples/Caddyfile /etc/caddy/Caddyfile

# Edit with your domain
sudo nano /etc/caddy/Caddyfile

# Reload Caddy
sudo systemctl reload caddy
```

### Nginx

```bash
# Install Nginx
sudo apt install nginx  # Debian/Ubuntu

# Copy the example config
sudo cp examples/nginx.conf /etc/nginx/sites-available/nopher

# Get TLS certificates (using certbot)
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d gemini.example.com

# Enable the site
sudo ln -s /etc/nginx/sites-available/nopher /etc/nginx/sites-enabled/

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

## Configuration

### Minimal Configuration

The bare minimum to get started:

```yaml
site:
  title: "My Nopher Instance"
  operator: "Your Name"

identity:
  npub: "npub1..."  # Your Nostr public key

protocols:
  gopher:
    enabled: true
    host: localhost
    port: 7070

  gemini:
    enabled: true
    host: localhost
    port: 11965
    tls:
      auto_generate: true

  finger:
    enabled: true
    port: 7079

relays:
  seeds:
    - wss://relay.damus.io
    - wss://nos.lol

storage:
  driver: sqlite
  sqlite_path: ./nopher.db
```

### Environment Variables

Sensitive values should be set via environment variables:

```bash
# Required: Your Nostr secret key
export NOPHER_NSEC="nsec1..."

# Optional: Redis URL for caching
export NOPHER_REDIS_URL="redis://localhost:6379"

# Optional: Log level override
export NOPHER_LOG_LEVEL="debug"
```

### File Locations

Different installation methods use different paths:

| Installation Method | Config Path | Data Path | Service User |
|---------------------|-------------|-----------|--------------|
| Package (DEB/RPM) | `/etc/nopher/nopher.yaml` | `/var/lib/nopher` | `nopher` |
| Homebrew | `/usr/local/etc/nopher.yaml` | `/usr/local/var/nopher` | current user |
| Docker | `/etc/nopher/nopher.yaml` (in container) | `/var/lib/nopher` (in container) | `nopher` |
| From Source | `./nopher.yaml` or `~/.config/nopher/nopher.yaml` | `./data` or specified | current user |

## Port Requirements

Nopher uses the following standard ports:

- **Port 70**: Gopher protocol (TCP)
- **Port 79**: Finger protocol (TCP)
- **Port 1965**: Gemini protocol (TCP with TLS)

### Running on Non-Privileged Ports

If you can't bind to ports < 1024:

```yaml
protocols:
  gopher:
    port: 7070  # Instead of 70

  gemini:
    port: 11965  # Instead of 1965

  finger:
    port: 7079  # Instead of 79
```

Then use a reverse proxy or `iptables` to forward traffic:

```bash
# Forward port 70 to 7070
sudo iptables -t nat -A PREROUTING -p tcp --dport 70 -j REDIRECT --to-port 7070
sudo iptables -t nat -A PREROUTING -p tcp --dport 79 -j REDIRECT --to-port 7079
sudo iptables -t nat -A PREROUTING -p tcp --dport 1965 -j REDIRECT --to-port 11965
```

## Verification

After installation, verify Nopher is running:

```bash
# Check if ports are listening
ss -tlnp | grep nopher

# Test Gopher (if enabled on port 70)
echo "" | nc localhost 70

# Test Finger (if enabled on port 79)
echo "user@localhost" | nc localhost 79

# Test Gemini requires a Gemini client
# gemget gemini://localhost/
```

## Troubleshooting

### Service won't start

```bash
# Check service status
sudo systemctl status nopher

# View full logs
sudo journalctl -u nopher -n 100 --no-pager

# Check config syntax
nopher --config /etc/nopher/nopher.yaml --validate
```

### Permission denied on ports < 1024

Either:
1. Run as root (not recommended)
2. Use non-privileged ports (7070, 7079, 11965) with reverse proxy
3. Grant CAP_NET_BIND_SERVICE capability:

```bash
sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/nopher
```

### TLS certificate issues

For Gemini with auto-generated certs:

```bash
# Check cert directory permissions
ls -la ~/.config/nopher/certs/

# Manually generate certs for testing
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout gemini.key -out gemini.crt \
  -days 365 -subj "/CN=localhost"
```

## Next Steps

- Read the [Configuration Guide](../memory/configuration.md)
- Learn about [Layouts and Sections](../memory/layouts_sections.md)
- See [Architecture Overview](../memory/architecture.md)

## Support

- GitHub Issues: https://github.com/sandwichfarm/nopher/issues
- Documentation: https://github.com/sandwichfarm/nopher
- Nostr: Contact the developer on Nostr!
