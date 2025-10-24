# nophr Installation Guide

This guide covers different ways to install and deploy nophr.

## Table of Contents

- [Quick Install](#quick-install)
- [Package Managers](#package-managers)
- [Docker](#docker)
- [From Source](#from-source)
- [Systemd Service](#systemd-service)
- [Reverse Proxy Setup](#reverse-proxy-setup)
- [Configuration](#configuration)

## Quick Install

The easiest way to install nophr is using our one-line installer:

```bash
curl -sSL https://nophr.io/install.sh | sh
```

This will:
- Detect your platform and architecture
- Download the latest release
- Install to `/usr/local/bin/nophr`
- Create example configuration at `~/.config/nophr/nophr.yaml`

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew tap sandwichfarm/tap
brew install nophr
```

### Debian/Ubuntu (DEB)

```bash
# Download the latest .deb package
curl -LO https://github.com/sandwichfarm/nophr/releases/download/v0.1.0/nophr_0.1.0_amd64.deb

# Install
sudo dpkg -i nophr_0.1.0_amd64.deb

# The post-install script will:
# - Create 'nophr' system user
# - Set up directories (/var/lib/nophr, /etc/nophr)
# - Install systemd service
```

### RHEL/CentOS/Fedora (RPM)

```bash
# Download the latest .rpm package
curl -LO https://github.com/sandwichfarm/nophr/releases/download/v0.1.0/nophr_0.1.0_amd64.rpm

# Install
sudo yum localinstall nophr_0.1.0_amd64.rpm
# or
sudo dnf install nophr_0.1.0_amd64.rpm
```

### Alpine (APK)

```bash
# Download the latest .apk package
curl -LO https://github.com/sandwichfarm/nophr/releases/download/v0.1.0/nophr_0.1.0_amd64.apk

# Install
sudo apk add --allow-untrusted nophr_0.1.0_amd64.apk
```

## Docker

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/sandwichfarm/nophr.git
cd nophr

# Copy example config
cp configs/nophr.example.yaml configs/nophr.yaml

# Edit your configuration
nano configs/nophr.yaml

# Set your NSEC (never commit this!)
export NOPHR_NSEC="nsec1..."

# Start the services
docker-compose up -d

# View logs
docker-compose logs -f nophr

# Stop the services
docker-compose down
```

### Using Docker directly

```bash
# Pull the latest image
docker pull ghcr.io/sandwichfarm/nophr:latest

# Run with config
docker run -d \
  --name nophr \
  -p 70:70 \
  -p 1965:1965 \
  -p 79:79 \
  -v ./nophr.yaml:/etc/nophr/nophr.yaml:ro \
  -v nophr-data:/var/lib/nophr \
  -e NOPHR_NSEC="nsec1..." \
  ghcr.io/sandwichfarm/nophr:latest
```

### Building from source with Docker

```bash
git clone https://github.com/sandwichfarm/nophr.git
cd nophr

# Build the image
docker build -t nophr:latest .

# Run
docker run -d \
  --name nophr \
  -p 70:70 \
  -p 1965:1965 \
  -p 79:79 \
  -v ./nophr.yaml:/etc/nophr/nophr.yaml:ro \
  -e NOPHR_NSEC="nsec1..." \
  nophr:latest
```

## From Source

### Prerequisites

- Go 1.25 or later
- Git

### Build and Install

```bash
# Clone the repository
git clone https://github.com/sandwichfarm/nophr.git
cd nophr

# Build
go build -o nophr ./cmd/nophr

# Install to /usr/local/bin
sudo mv nophr /usr/local/bin/

# Or keep it local
./nophr --help
```

### Development Mode

```bash
# Run directly with Go
go run ./cmd/nophr --config ./test-config.yaml

# Build for development
go build -o nophr ./cmd/nophr
./nophr --config ./test-config.yaml
```

## Systemd Service

After installing via package manager, enable and start the service:

```bash
# Create your config
sudo cp /etc/nophr/nophr.example.yaml /etc/nophr/nophr.yaml
sudo nano /etc/nophr/nophr.yaml

# Set your NSEC (secure method)
echo 'NOPHR_NSEC="nsec1..."' | sudo tee /etc/default/nophr

# Enable and start service
sudo systemctl enable nophr
sudo systemctl start nophr

# Check status
sudo systemctl status nophr

# View logs
sudo journalctl -u nophr -f
```

### Manual Systemd Setup

If you installed from source or via the installer script:

```bash
# Copy the service file
sudo cp scripts/systemd/nophr.service /etc/systemd/system/

# Create nophr user
sudo useradd --system --no-create-home --shell /bin/false nophr

# Create directories
sudo mkdir -p /var/lib/nophr /etc/nophr
sudo chown nophr:nophr /var/lib/nophr

# Copy config
sudo cp configs/nophr.example.yaml /etc/nophr/nophr.yaml

# Edit config
sudo nano /etc/nophr/nophr.yaml

# Reload systemd
sudo systemctl daemon-reload

# Enable and start
sudo systemctl enable nophr
sudo systemctl start nophr
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
sudo cp examples/nginx.conf /etc/nginx/sites-available/nophr

# Get TLS certificates (using certbot)
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d gemini.example.com

# Enable the site
sudo ln -s /etc/nginx/sites-available/nophr /etc/nginx/sites-enabled/

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

## Configuration

### Minimal Configuration

The bare minimum to get started:

```yaml
site:
  title: "My nophr Instance"
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
  sqlite_path: ./nophr.db
```

### Environment Variables

Sensitive values should be set via environment variables:

```bash
# Required: Your Nostr secret key
export NOPHR_NSEC="nsec1..."

# Optional: Redis URL for caching
export NOPHR_REDIS_URL="redis://localhost:6379"

# Optional: Log level override
export NOPHR_LOG_LEVEL="debug"
```

### File Locations

Different installation methods use different paths:

| Installation Method | Config Path | Data Path | Service User |
|---------------------|-------------|-----------|--------------|
| Package (DEB/RPM) | `/etc/nophr/nophr.yaml` | `/var/lib/nophr` | `nophr` |
| Homebrew | `/usr/local/etc/nophr.yaml` | `/usr/local/var/nophr` | current user |
| Docker | `/etc/nophr/nophr.yaml` (in container) | `/var/lib/nophr` (in container) | `nophr` |
| From Source | `./nophr.yaml` or `~/.config/nophr/nophr.yaml` | `./data` or specified | current user |

## Port Requirements

nophr uses the following standard ports:

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

After installation, verify nophr is running:

```bash
# Check if ports are listening
ss -tlnp | grep nophr

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
sudo systemctl status nophr

# View full logs
sudo journalctl -u nophr -n 100 --no-pager

# Check config syntax
nophr --config /etc/nophr/nophr.yaml --validate
```

### Permission denied on ports < 1024

Either:
1. Run as root (not recommended)
2. Use non-privileged ports (7070, 7079, 11965) with reverse proxy
3. Grant CAP_NET_BIND_SERVICE capability:

```bash
sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/nophr
```

### TLS certificate issues

For Gemini with auto-generated certs:

```bash
# Check cert directory permissions
ls -la ~/.config/nophr/certs/

# Manually generate certs for testing
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout gemini.key -out gemini.crt \
  -days 365 -subj "/CN=localhost"
```

## Next Steps

- Read the [Configuration Guide](configuration.md)
- Learn about [Sections](configuration.md#sections)
- See [Architecture Overview](architecture.md)

## Support

- GitHub Issues: https://github.com/sandwichfarm/nophr/issues
- Documentation: https://github.com/sandwichfarm/nophr
- Nostr: Contact the developer on Nostr!
