#!/bin/sh
# One-line installer for Nopher
# Usage: curl -sSL https://nophr.io/install.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        FreeBSD*)
            OS="freebsd"
            ;;
        OpenBSD*)
            OS="openbsd"
            ;;
        *)
            echo "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv7)
            ARCH="arm"
            ;;
        *)
            echo "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    echo "${GREEN}Detected platform: ${OS}_${ARCH}${NC}"
}

# Get latest release version
get_latest_version() {
    echo "${YELLOW}Fetching latest version...${NC}"
    VERSION=$(curl -sSL https://api.github.com/repos/sandwichfarm/nophr/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        echo "${RED}Failed to get latest version${NC}"
        exit 1
    fi

    echo "${GREEN}Latest version: v${VERSION}${NC}"
}

# Download and install binary
install_binary() {
    BINARY_NAME="nophr_${VERSION}_${OS}_${ARCH}"
    if [ "$OS" = "linux" ]; then
        ARCHIVE_NAME="${BINARY_NAME}.tar.gz"
    else
        ARCHIVE_NAME="${BINARY_NAME}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/sandwichfarm/nophr/releases/download/v${VERSION}/${ARCHIVE_NAME}"

    echo "${YELLOW}Downloading ${ARCHIVE_NAME}...${NC}"
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    if ! curl -sSL "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"; then
        echo "${RED}Failed to download Nopher${NC}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    echo "${YELLOW}Extracting...${NC}"
    tar -xzf "$ARCHIVE_NAME"

    # Determine install location
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        SUDO=""
    elif [ "$(id -u)" -eq 0 ]; then
        INSTALL_DIR="/usr/local/bin"
        SUDO=""
    else
        echo "${YELLOW}Need sudo to install to /usr/local/bin${NC}"
        INSTALL_DIR="/usr/local/bin"
        SUDO="sudo"
    fi

    echo "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"
    $SUDO mv nophr "$INSTALL_DIR/nophr"
    $SUDO chmod +x "$INSTALL_DIR/nophr"

    cd - > /dev/null
    rm -rf "$TEMP_DIR"

    echo "${GREEN}Nopher installed successfully!${NC}"
}

# Install via package manager (optional)
install_via_package_manager() {
    if [ "$OS" = "linux" ]; then
        if command -v apt-get > /dev/null 2>&1; then
            echo "${YELLOW}Detected Debian/Ubuntu. You can also install via apt:${NC}"
            echo "  curl -sSL https://github.com/sandwichfarm/nophr/releases/download/v${VERSION}/nophr_${VERSION}_amd64.deb -o nophr.deb"
            echo "  sudo dpkg -i nophr.deb"
            echo ""
        elif command -v yum > /dev/null 2>&1; then
            echo "${YELLOW}Detected RHEL/CentOS. You can also install via yum:${NC}"
            echo "  curl -sSL https://github.com/sandwichfarm/nophr/releases/download/v${VERSION}/nophr_${VERSION}_amd64.rpm -o nophr.rpm"
            echo "  sudo yum localinstall nophr.rpm"
            echo ""
        fi
    elif [ "$OS" = "darwin" ]; then
        if command -v brew > /dev/null 2>&1; then
            echo "${YELLOW}Detected Homebrew. You can also install via brew:${NC}"
            echo "  brew install sandwichfarm/tap/nophr"
            echo ""
        fi
    fi
}

# Create config directory
setup_config() {
    CONFIG_DIR="$HOME/.config/nophr"

    if [ ! -d "$CONFIG_DIR" ]; then
        echo "${YELLOW}Creating config directory...${NC}"
        mkdir -p "$CONFIG_DIR"
    fi

    if [ ! -f "$CONFIG_DIR/nophr.yaml" ]; then
        echo "${YELLOW}Creating example config...${NC}"
        cat > "$CONFIG_DIR/nophr.yaml" <<'EOF'
# Nopher Configuration

logging:
  level: info
  format: text

site:
  title: "My Nopher Instance"
  operator: "Your Name"
  description: "Nostr to Gopher/Gemini/Finger Gateway"

identity:
  npub: ""  # Your Nostr public key
  nsec: ""  # Your Nostr secret key (optional)

storage:
  driver: sqlite
  sqlite_path: ~/.local/share/nophr/nophr.db
  lmdb_path: ~/.local/share/nophr/lmdb
  lmdb_max_size_mb: 1024

protocols:
  gopher:
    enabled: true
    host: localhost
    port: 7070
    bind: "0.0.0.0"

  gemini:
    enabled: true
    host: localhost
    port: 11965
    bind: "0.0.0.0"
    tls:
      cert_path: ~/.config/nophr/certs/gemini.crt
      key_path: ~/.config/nophr/certs/gemini.key
      auto_generate: true

  finger:
    enabled: true
    port: 7079
    bind: "0.0.0.0"
    max_users: 100

relays:
  seeds:
    - wss://relay.damus.io
    - wss://nos.lol
  discovery:
    enabled: true

sync:
  enabled: true
  scope:
    mode: self
  interval_minutes: 60

inbox:
  enabled: true
  include_replies: true
  include_mentions: true
  include_reactions: true
  include_zaps: true
  include_reposts: false
EOF

        echo "${GREEN}Created config at ${CONFIG_DIR}/nophr.yaml${NC}"
        echo "${YELLOW}Please edit this file to configure your Nopher instance${NC}"
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    echo "${GREEN}==================================================${NC}"
    echo "${GREEN}Nopher has been installed successfully!${NC}"
    echo "${GREEN}==================================================${NC}"
    echo ""
    echo "Quick start:"
    echo "  1. Edit your config: $HOME/.config/nophr/nophr.yaml"
    echo "  2. Run Nopher: nophr --config $HOME/.config/nophr/nophr.yaml"
    echo ""
    echo "For production deployment:"
    echo "  - Systemd: https://github.com/sandwichfarm/nophr/tree/main/scripts/systemd"
    echo "  - Docker: https://github.com/sandwichfarm/nophr#docker"
    echo "  - Reverse Proxy: https://github.com/sandwichfarm/nophr/tree/main/examples"
    echo ""
    echo "Documentation: https://github.com/sandwichfarm/nophr"
    echo "Report issues: https://github.com/sandwichfarm/nophr/issues"
    echo ""
}

# Main installation flow
main() {
    echo "${GREEN}==================================================${NC}"
    echo "${GREEN}Nopher Installer${NC}"
    echo "${GREEN}==================================================${NC}"
    echo ""

    detect_platform
    get_latest_version
    install_binary
    install_via_package_manager
    setup_config
    print_usage
}

main
