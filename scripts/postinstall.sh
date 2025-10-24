#!/bin/sh
# Post-install script for nophr package

# Create nophr user if it doesn't exist
if ! id -u nophr >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false nophr
fi

# Create data directory
mkdir -p /var/lib/nophr
chown nophr:nophr /var/lib/nophr
chmod 750 /var/lib/nophr

# Create config directory
mkdir -p /etc/nophr
chmod 755 /etc/nophr

# Reload systemd if available
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    echo "Nopher installed. Enable and start with:"
    echo "  sudo systemctl enable nophr"
    echo "  sudo systemctl start nophr"
fi

echo ""
echo "Configure nophr by editing /etc/nophr/nophr.yaml"
echo "See /etc/nophr/nophr.example.yaml for reference"
