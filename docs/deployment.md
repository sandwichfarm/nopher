# Deployment Guide

**Status:** Production deployment guide

Complete guide to deploying Nopher in production: system configuration, port binding, TLS certificates, systemd services, and monitoring.

## Prerequisites

- Linux server (Ubuntu 20.04+, Debian 11+, or similar)
- Root/sudo access
- Public IP or domain name
- Open ports: 70 (Gopher), 79 (Finger), 1965 (Gemini)

---

## Table of Contents

- [System Requirements](#system-requirements)
- [Port Binding](#port-binding)
- [TLS Certificates](#tls-certificates)
- [Systemd Service](#systemd-service)
- [Reverse Proxy](#reverse-proxy)
- [Monitoring](#monitoring)
- [Backups](#backups)
- [Updates](#updates)

---

## System Requirements

### Minimum

- **CPU:** 1 core
- **RAM:** 256MB
- **Disk:** 1GB (for binary + database)
- **Network:** Stable internet, low latency to Nostr relays

### Recommended

- **CPU:** 2 cores
- **RAM:** 512MB-1GB
- **Disk:** 5-10GB (for database growth)
- **Network:** <100ms latency to major Nostr relays

### Scaling

**For <100K events:**
- SQLite backend
- 512MB RAM
- 1 core

**For >100K events:**
- LMDB backend
- 1GB RAM
- 2 cores

---

## Port Binding

Ports below 1024 require root privileges. Nopher needs ports 70, 79, and optionally 1965 (usually OK).

### Option 1: systemd Socket Activation (Recommended)

Systemd can bind ports as root, then pass sockets to unprivileged Nopher process.

**Create socket units:**

`/etc/systemd/system/nopher-gopher.socket`:
```ini
[Unit]
Description=Nopher Gopher Socket
PartOf=nopher.service

[Socket]
ListenStream=0.0.0.0:70
Accept=no

[Install]
WantedBy=sockets.target
```

`/etc/systemd/system/nopher-finger.socket`:
```ini
[Unit]
Description=Nopher Finger Socket
PartOf=nopher.service

[Socket]
ListenStream=0.0.0.0:79
Accept=no

[Install]
WantedBy=sockets.target
```

`/etc/systemd/system/nopher-gemini.socket`:
```ini
[Unit]
Description=Nopher Gemini Socket
PartOf=nopher.service

[Socket]
ListenStream=0.0.0.0:1965
Accept=no

[Install]
WantedBy=sockets.target
```

**Update service unit:**

`/etc/systemd/system/nopher.service` (add `Requires` and `After`):
```ini
[Unit]
Description=Nopher - Nostr to Gopher/Gemini/Finger Gateway
After=network.target nopher-gopher.socket nopher-finger.socket nopher-gemini.socket
Requires=nopher-gopher.socket nopher-finger.socket nopher-gemini.socket

[Service]
Type=simple
User=nopher
Group=nopher
WorkingDirectory=/opt/nopher
ExecStart=/usr/local/bin/nopher --config /opt/nopher/nopher.yaml
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

**Enable and start:**
```bash
sudo systemctl enable nopher-gopher.socket nopher-finger.socket nopher-gemini.socket
sudo systemctl start nopher-gopher.socket nopher-finger.socket nopher-gemini.socket
sudo systemctl enable nopher.service
sudo systemctl start nopher.service
```

**Verify:**
```bash
sudo systemctl status nopher
sudo ss -tlnp | grep -E ':(70|79|1965)'
```

### Option 2: Port Forwarding (iptables)

Forward high ports → low ports, run Nopher on high ports.

**Configure Nopher on high ports:**
```yaml
protocols:
  gopher:
    port: 7070
  finger:
    port: 7979
  gemini:
    port: 11965
```

**Forward ports:**
```bash
# Gopher: 70 → 7070
sudo iptables -t nat -A PREROUTING -p tcp --dport 70 -j REDIRECT --to-port 7070

# Finger: 79 → 7979
sudo iptables -t nat -A PREROUTING -p tcp --dport 79 -j REDIRECT --to-port 7979

# Gemini: 1965 → 11965
sudo iptables -t nat -A PREROUTING -p tcp --dport 1965 -j REDIRECT --to-port 11965

# Save rules
sudo iptables-save > /etc/iptables/rules.v4
```

**Make persistent:**
```bash
# Install iptables-persistent
sudo apt install iptables-persistent

# Or on systemd-based systems, create service
sudo systemctl enable netfilter-persistent
sudo systemctl start netfilter-persistent
```

### Option 3: Run as Root (Not Recommended)

**Only for testing!** Do not run production as root.

```bash
sudo /usr/local/bin/nopher --config /opt/nopher/nopher.yaml
```

---

## TLS Certificates

Gemini requires TLS. Two options: self-signed (personal use) or Let's Encrypt (production).

### Self-Signed Certificates

**Automatic (recommended):**

```yaml
protocols:
  gemini:
    tls:
      auto_generate: true
```

Nopher generates self-signed cert on first run.

**Manual generation:**
```bash
mkdir -p /opt/nopher/certs
openssl req -x509 -newkey rsa:4096 \
  -keyout /opt/nopher/certs/key.pem \
  -out /opt/nopher/certs/cert.pem \
  -days 365 -nodes \
  -subj "/CN=gemini.example.com"

chown nopher:nopher /opt/nopher/certs/*.pem
chmod 600 /opt/nopher/certs/key.pem
```

**Configuration:**
```yaml
protocols:
  gemini:
    tls:
      cert_path: "/opt/nopher/certs/cert.pem"
      key_path: "/opt/nopher/certs/key.pem"
      auto_generate: false
```

**Note:** Self-signed certs require TOFU (Trust On First Use) in Gemini clients.

### Let's Encrypt Certificates

**Using certbot:**

```bash
# Install certbot
sudo apt install certbot

# Get certificate (HTTP-01 challenge)
# Note: Requires port 80 open temporarily
sudo certbot certonly --standalone -d gemini.example.com

# Certificates saved to:
# /etc/letsencrypt/live/gemini.example.com/fullchain.pem
# /etc/letsencrypt/live/gemini.example.com/privkey.pem
```

**Copy to Nopher:**
```bash
sudo cp /etc/letsencrypt/live/gemini.example.com/fullchain.pem /opt/nopher/certs/cert.pem
sudo cp /etc/letsencrypt/live/gemini.example.com/privkey.pem /opt/nopher/certs/key.pem
sudo chown nopher:nopher /opt/nopher/certs/*.pem
sudo chmod 600 /opt/nopher/certs/key.pem
```

**Configuration:**
```yaml
protocols:
  gemini:
    tls:
      cert_path: "/opt/nopher/certs/cert.pem"
      key_path: "/opt/nopher/certs/key.pem"
      auto_generate: false
```

**Auto-renewal:**

Certbot creates cron job automatically. Add post-renewal hook:

`/etc/letsencrypt/renewal-hooks/post/nopher-reload.sh`:
```bash
#!/bin/bash
cp /etc/letsencrypt/live/gemini.example.com/fullchain.pem /opt/nopher/certs/cert.pem
cp /etc/letsencrypt/live/gemini.example.com/privkey.pem /opt/nopher/certs/key.pem
chown nopher:nopher /opt/nopher/certs/*.pem
chmod 600 /opt/nopher/certs/key.pem
systemctl restart nopher
```

```bash
sudo chmod +x /etc/letsencrypt/renewal-hooks/post/nopher-reload.sh
```

---

## Systemd Service

Run Nopher as a systemd service for automatic startup and management.

### Create User

```bash
sudo useradd --system --create-home --home-dir /opt/nopher --shell /bin/bash nopher
```

### Install Binary

```bash
sudo cp dist/nopher /usr/local/bin/nopher
sudo chmod +x /usr/local/bin/nopher
```

### Create Configuration

```bash
sudo mkdir -p /opt/nopher/data
sudo mkdir -p /opt/nopher/certs
sudo cp nopher.yaml /opt/nopher/nopher.yaml
sudo chown -R nopher:nopher /opt/nopher
```

### Create Service Unit

`/etc/systemd/system/nopher.service`:
```ini
[Unit]
Description=Nopher - Nostr to Gopher/Gemini/Finger Gateway
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=nopher
Group=nopher
WorkingDirectory=/opt/nopher

# Main command
ExecStart=/usr/local/bin/nopher --config /opt/nopher/nopher.yaml

# Restart policy
Restart=on-failure
RestartSec=10s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/nopher

# Resource limits
LimitNOFILE=65536

# Environment
Environment="NOPHER_NSEC_FILE=/opt/nopher/nsec"

[Install]
WantedBy=multi-user.target
```

**Store nsec securely:**
```bash
echo "nsec1..." | sudo tee /opt/nopher/nsec
sudo chmod 600 /opt/nopher/nsec
sudo chown nopher:nopher /opt/nopher/nsec
```

### Enable and Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable nopher.service
sudo systemctl start nopher.service
```

### Check Status

```bash
sudo systemctl status nopher
sudo journalctl -u nopher -f
```

### Manage Service

```bash
# Start
sudo systemctl start nopher

# Stop
sudo systemctl stop nopher

# Restart
sudo systemctl restart nopher

# Reload config (if supported)
sudo systemctl reload nopher

# View logs
sudo journalctl -u nopher -n 100
```

---

## Reverse Proxy

For advanced setups, you can run Nopher behind a reverse proxy (though protocols are non-HTTP).

### Gopher Proxy (socat)

Forward Gopher through socat:

```bash
sudo apt install socat

# Forward external :70 → Nopher :7070
socat TCP4-LISTEN:70,fork TCP4:localhost:7070
```

**As systemd service:**

`/etc/systemd/system/gopher-proxy.service`:
```ini
[Unit]
Description=Gopher Proxy
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/socat TCP4-LISTEN:70,fork TCP4:localhost:7070
Restart=always

[Install]
WantedBy=multi-user.target
```

### Gemini Proxy (stunnel)

Terminate TLS with stunnel, forward to Nopher:

**Install stunnel:**
```bash
sudo apt install stunnel4
```

**Configure:**

`/etc/stunnel/nopher.conf`:
```ini
[gemini]
accept = 1965
connect = localhost:11965
cert = /opt/nopher/certs/cert.pem
key = /opt/nopher/certs/key.pem
```

**Enable:**
```bash
sudo systemctl enable stunnel4
sudo systemctl start stunnel4
```

**Configure Nopher (no TLS):**
```yaml
protocols:
  gemini:
    port: 11965
    tls:
      auto_generate: false
```

---

## Firewall

Ensure required ports are open.

### ufw (Ubuntu/Debian)

```bash
# Enable firewall
sudo ufw enable

# SSH (important!)
sudo ufw allow 22/tcp

# Nopher ports
sudo ufw allow 70/tcp    # Gopher
sudo ufw allow 79/tcp    # Finger
sudo ufw allow 1965/tcp  # Gemini

# Check status
sudo ufw status
```

### iptables

```bash
# Allow SSH
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Allow Nopher
sudo iptables -A INPUT -p tcp --dport 70 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 79 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 1965 -j ACCEPT

# Drop others
sudo iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
sudo iptables -A INPUT -j DROP

# Save
sudo iptables-save > /etc/iptables/rules.v4
```

### firewalld (RHEL/CentOS)

```bash
sudo firewall-cmd --permanent --add-port=70/tcp
sudo firewall-cmd --permanent --add-port=79/tcp
sudo firewall-cmd --permanent --add-port=1965/tcp
sudo firewall-cmd --reload
```

---

## Monitoring

### System Logs

**View logs:**
```bash
sudo journalctl -u nopher -f
```

**Recent errors:**
```bash
sudo journalctl -u nopher -p err -n 50
```

### Health Checks

**Check if Nopher is running:**
```bash
sudo systemctl is-active nopher
```

**Check ports:**
```bash
sudo ss -tlnp | grep -E ':(70|79|1965)'
```

**Expected output:**
```
LISTEN  0  128  0.0.0.0:70    0.0.0.0:*  users:(("nopher",pid=1234,fd=3))
LISTEN  0  128  0.0.0.0:79    0.0.0.0:*  users:(("nopher",pid=1234,fd=4))
LISTEN  0  128  0.0.0.0:1965  0.0.0.0:*  users:(("nopher",pid=1234,fd=5))
```

### Protocol Tests

**Gopher:**
```bash
echo "/" | nc localhost 70
```

**Gemini:**
```bash
echo "gemini://localhost/" | openssl s_client -connect localhost:1965 -quiet
```

**Finger:**
```bash
echo "" | nc localhost 79
```

### Database Size

```bash
du -h /opt/nopher/data/nopher.db
```

### Event Count

```bash
sqlite3 /opt/nopher/data/nopher.db "SELECT COUNT(*) FROM events;"
```

### Automated Monitoring (cron)

**Create monitoring script:**

`/opt/nopher/scripts/health-check.sh`:
```bash
#!/bin/bash
# Check if Nopher is running
if ! systemctl is-active --quiet nopher; then
    echo "Nopher is down! Restarting..."
    systemctl restart nopher
    echo "Nopher restarted at $(date)" >> /var/log/nopher-health.log
fi

# Check database size
DB_SIZE=$(du -m /opt/nopher/data/nopher.db | cut -f1)
if [ "$DB_SIZE" -gt 10000 ]; then
    echo "Database size exceeded 10GB: ${DB_SIZE}MB" >> /var/log/nopher-health.log
fi
```

**Add to cron:**
```bash
sudo chmod +x /opt/nopher/scripts/health-check.sh
sudo crontab -e
# Add line:
*/5 * * * * /opt/nopher/scripts/health-check.sh
```

---

## Backups

### Automated Backups

**Backup script:**

`/opt/nopher/scripts/backup.sh`:
```bash
#!/bin/bash
BACKUP_DIR="/var/backups/nopher"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup database
cp /opt/nopher/data/nopher.db "$BACKUP_DIR/nopher-$DATE.db"

# Backup config
cp /opt/nopher/nopher.yaml "$BACKUP_DIR/nopher-$DATE.yaml"

# Keep last 7 days
find "$BACKUP_DIR" -name "nopher-*.db" -mtime +7 -delete

echo "Backup completed: $DATE"
```

**Schedule daily backups:**
```bash
sudo chmod +x /opt/nopher/scripts/backup.sh
sudo crontab -e
# Add:
0 2 * * * /opt/nopher/scripts/backup.sh >> /var/log/nopher-backup.log 2>&1
```

### Off-site Backups

**rsync to remote server:**
```bash
rsync -avz /var/backups/nopher/ user@backup-server:/backups/nopher/
```

**Or use cloud storage (rclone):**
```bash
sudo apt install rclone
rclone sync /var/backups/nopher/ remote:nopher-backups/
```

---

## Updates

### Update Nopher

**Stop service:**
```bash
sudo systemctl stop nopher
```

**Backup current binary:**
```bash
sudo cp /usr/local/bin/nopher /usr/local/bin/nopher.bak
```

**Install new binary:**
```bash
sudo cp dist/nopher /usr/local/bin/nopher
sudo chmod +x /usr/local/bin/nopher
```

**Start service:**
```bash
sudo systemctl start nopher
```

**Verify:**
```bash
sudo systemctl status nopher
/usr/local/bin/nopher --version
```

### Rollback

If update causes issues:

```bash
sudo systemctl stop nopher
sudo cp /usr/local/bin/nopher.bak /usr/local/bin/nopher
sudo systemctl start nopher
```

---

## Performance Tuning

### File Descriptors

Increase file descriptor limit for high traffic:

`/etc/security/limits.conf`:
```
nopher soft nofile 65536
nopher hard nofile 65536
```

Or in systemd service:
```ini
[Service]
LimitNOFILE=65536
```

### Database Tuning

**SQLite WAL mode** (already enabled by Khatru):
```bash
sqlite3 /opt/nopher/data/nopher.db "PRAGMA journal_mode=WAL;"
```

**Vacuum periodically:**
```bash
# Weekly cron
0 3 * * 0 sqlite3 /opt/nopher/data/nopher.db "VACUUM;"
```

### LMDB Max Size

For high-volume deployments:

```yaml
storage:
  driver: "lmdb"
  lmdb_max_size_mb: 20480  # 20GB
```

---

## Troubleshooting Deployment

### "Permission denied" binding to port

**Cause:** Ports <1024 require root.

**Fix:** Use systemd socket activation or port forwarding.

### "Address already in use"

**Cause:** Another service using port.

**Fix:**
```bash
# Find process
sudo ss -tlnp | grep :70

# Kill process or change port
```

### "TLS handshake failed" (Gemini)

**Cause:** Invalid certificate or wrong path.

**Fix:**
- Check cert paths in config
- Verify cert files exist and have correct permissions
- Regenerate certificate

### Database locked

**Cause:** Multiple Nopher instances or unclean shutdown.

**Fix:**
```bash
# Ensure only one instance
sudo systemctl stop nopher
# Check for stale locks
rm -f /opt/nopher/data/nopher.db-shm /opt/nopher/data/nopher.db-wal
sudo systemctl start nopher
```

### Service crashes on start

**Check logs:**
```bash
sudo journalctl -u nopher -n 100
```

**Common causes:**
- Invalid config (missing npub, bad relay URLs)
- Missing data directory
- Permission issues

---

## Security Checklist

- [ ] Run as non-root user (`nopher`)
- [ ] Use systemd socket activation or port forwarding
- [ ] Store nsec in separate file with 600 permissions
- [ ] Enable firewall (ufw/iptables)
- [ ] Use proper TLS certs (Let's Encrypt)
- [ ] Set up automated backups
- [ ] Monitor logs regularly
- [ ] Keep Nopher updated
- [ ] Limit database size (retention policy)

---

**Next:** [Troubleshooting Guide](troubleshooting.md) | [Configuration Reference](configuration.md) | [Getting Started](getting-started.md)
