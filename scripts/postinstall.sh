#!/bin/sh
# Post-install script for nopher package

# Create nopher user if it doesn't exist
if ! id -u nopher >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false nopher
fi

# Create data directory
mkdir -p /var/lib/nopher
chown nopher:nopher /var/lib/nopher
chmod 750 /var/lib/nopher

# Create config directory
mkdir -p /etc/nopher
chmod 755 /etc/nopher

# Reload systemd if available
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    echo "Nopher installed. Enable and start with:"
    echo "  sudo systemctl enable nopher"
    echo "  sudo systemctl start nopher"
fi

echo ""
echo "Configure nopher by editing /etc/nopher/nopher.yaml"
echo "See /etc/nopher/nopher.example.yaml for reference"
