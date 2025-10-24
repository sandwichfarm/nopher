#!/usr/bin/env bash
# Post-install script for system packages (APT/RPM)

set -euo pipefail

# Create nopher user if it doesn't exist
if ! id -u nopher &>/dev/null; then
    useradd --system --user-group --no-create-home --shell /bin/false nopher
fi

# Create data directory
mkdir -p /var/lib/nopher
chown nopher:nopher /var/lib/nopher
chmod 750 /var/lib/nopher

# Create config directory
mkdir -p /etc/nopher
chmod 755 /etc/nopher

# Copy example config if main config doesn't exist
if [ ! -f /etc/nopher/nopher.yaml ] && [ -f /etc/nopher/nopher.example.yaml ]; then
    cp /etc/nopher/nopher.example.yaml /etc/nopher/nopher.yaml
    chmod 640 /etc/nopher/nopher.yaml
    chown root:nopher /etc/nopher/nopher.yaml
    echo "Created /etc/nopher/nopher.yaml - please edit before starting"
fi

# Reload systemd
if command -v systemctl &>/dev/null; then
    systemctl daemon-reload
fi

echo "Nopher installed successfully!"
echo "Edit /etc/nopher/nopher.yaml and run: systemctl start nopher"
