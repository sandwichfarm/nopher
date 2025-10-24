package sync

import (
	"context"
	"time"
)

// RelayInfo contains information about a relay
type RelayInfo struct {
	url         string
	connected   bool
	lastConnect *time.Time
	lastError   error
}

// URL returns the relay URL
func (r *RelayInfo) URL() string {
	return r.url
}

// IsConnected returns whether the relay is currently connected
func (r *RelayInfo) IsConnected() bool {
	return r.connected
}

// LastConnectTime returns the last time the relay connected
func (r *RelayInfo) LastConnectTime() *time.Time {
	return r.lastConnect
}

// LastError returns the last error from the relay
func (r *RelayInfo) LastError() error {
	return r.lastError
}

// GetRelays returns information about all configured relays
func (e *Engine) GetRelays() []*RelayInfo {
	// Get relay URLs from discovery/client
	relays := e.discovery.GetRelays()

	infos := make([]*RelayInfo, 0, len(relays))
	for _, relay := range relays {
		info := &RelayInfo{
			url:       relay.URL,
			connected: relay.Connected,
		}

		if relay.LastConnect != nil {
			info.lastConnect = relay.LastConnect
		}

		if relay.LastError != nil {
			info.lastError = relay.LastError
		}

		infos = append(infos, info)
	}

	return infos
}

// TotalSynced returns the total number of events synced
func (e *Engine) TotalSynced(ctx context.Context) (int64, error) {
	// Count all events in storage
	return e.storage.CountEvents(ctx)
}

// LastSyncTime returns the last time any event was synced
func (e *Engine) LastSyncTime(ctx context.Context) (*time.Time, error) {
	// Get the newest event timestamp from storage
	_, newest, err := e.storage.EventTimeRange(ctx)
	if err != nil {
		return nil, err
	}

	return newest, nil
}
