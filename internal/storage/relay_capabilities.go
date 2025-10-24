package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RelayCapabilities tracks what features a relay supports
type RelayCapabilities struct {
	URL                string
	SupportsNegentropy bool
	NIP11Software      string // Software name from NIP-11
	NIP11Version       string // Version from NIP-11
	LastChecked        time.Time
	CheckExpiry        time.Time // When to re-check capabilities (7 days from last check)
}

// GetRelayCapabilities retrieves cached capability information for a relay
func (s *Storage) GetRelayCapabilities(ctx context.Context, url string) (*RelayCapabilities, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT url, supports_negentropy, nip11_software, nip11_version,
		       last_checked, check_expiry
		FROM relay_capabilities
		WHERE url = ?
	`, url)

	var caps RelayCapabilities
	var lastChecked, checkExpiry int64

	err := row.Scan(
		&caps.URL,
		&caps.SupportsNegentropy,
		&caps.NIP11Software,
		&caps.NIP11Version,
		&lastChecked,
		&checkExpiry,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No cached data
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query relay capabilities: %w", err)
	}

	caps.LastChecked = time.Unix(lastChecked, 0)
	caps.CheckExpiry = time.Unix(checkExpiry, 0)

	return &caps, nil
}

// SaveRelayCapabilities stores capability information for a relay
func (s *Storage) SaveRelayCapabilities(ctx context.Context, caps *RelayCapabilities) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO relay_capabilities (
			url, supports_negentropy, nip11_software, nip11_version,
			last_checked, check_expiry
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			supports_negentropy = excluded.supports_negentropy,
			nip11_software = excluded.nip11_software,
			nip11_version = excluded.nip11_version,
			last_checked = excluded.last_checked,
			check_expiry = excluded.check_expiry
	`,
		caps.URL,
		caps.SupportsNegentropy,
		caps.NIP11Software,
		caps.NIP11Version,
		caps.LastChecked.Unix(),
		caps.CheckExpiry.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to save relay capabilities: %w", err)
	}

	return nil
}

// IsRelayCapabilityExpired checks if cached capability data needs refresh
func (s *Storage) IsRelayCapabilityExpired(ctx context.Context, url string) (bool, error) {
	caps, err := s.GetRelayCapabilities(ctx, url)
	if err != nil {
		return false, err
	}
	if caps == nil {
		return true, nil // No cached data = expired
	}

	return time.Now().After(caps.CheckExpiry), nil
}
