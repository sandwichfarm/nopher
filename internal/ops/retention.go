package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/sandwich/nopher/internal/config"
	"github.com/sandwich/nopher/internal/storage"
)

// RetentionManager handles data retention and pruning
type RetentionManager struct {
	storage *storage.Storage
	config  *config.Retention
	logger  *Logger
}

// NewRetentionManager creates a new retention manager
func NewRetentionManager(st *storage.Storage, cfg *config.Retention, logger *Logger) *RetentionManager {
	return &RetentionManager{
		storage: st,
		config:  cfg,
		logger:  logger.WithComponent("retention"),
	}
}

// PruneOldEvents deletes events older than the retention period
func (r *RetentionManager) PruneOldEvents(ctx context.Context) (int64, error) {
	start := time.Now()

	// Calculate cutoff time
	cutoff := time.Now().AddDate(0, 0, -r.config.KeepDays)

	r.logger.Info("starting retention pruning",
		"cutoff", cutoff.Format(time.RFC3339),
		"keep_days", r.config.KeepDays)

	// Delete events before cutoff
	deleted, err := r.storage.DeleteEventsBefore(ctx, cutoff)
	if err != nil {
		r.logger.LogRetentionPrune(int(deleted), time.Since(start), err)
		return 0, fmt.Errorf("failed to prune old events: %w", err)
	}

	r.logger.LogRetentionPrune(int(deleted), time.Since(start), nil)
	return deleted, nil
}

// PruneByKind deletes all events of a specific kind
func (r *RetentionManager) PruneByKind(ctx context.Context, kind int) (int64, error) {
	start := time.Now()

	r.logger.Info("pruning events by kind", "kind", kind)

	deleted, err := r.storage.DeleteEventsByKind(ctx, kind)
	if err != nil {
		r.logger.LogRetentionPrune(int(deleted), time.Since(start), err)
		return 0, fmt.Errorf("failed to prune events by kind: %w", err)
	}

	r.logger.Info("pruned events by kind",
		"kind", kind,
		"deleted", deleted,
		"duration_ms", time.Since(start).Milliseconds())

	return deleted, nil
}

// ShouldPruneOnStart returns true if pruning should run on startup
func (r *RetentionManager) ShouldPruneOnStart() bool {
	return r.config.PruneOnStart
}

// GetRetentionStats returns statistics about retention
func (r *RetentionManager) GetRetentionStats(ctx context.Context) (*RetentionStats, error) {
	stats := &RetentionStats{
		KeepDays:     r.config.KeepDays,
		PruneOnStart: r.config.PruneOnStart,
	}

	// Get total events
	total, err := r.storage.CountEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}
	stats.TotalEvents = total

	// Get time range
	oldest, newest, err := r.storage.EventTimeRange(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get event time range: %w", err)
	}

	if oldest != nil {
		stats.OldestEvent = *oldest
	}
	if newest != nil {
		stats.NewestEvent = *newest
	}

	// Calculate events eligible for pruning
	cutoff := time.Now().AddDate(0, 0, -r.config.KeepDays)
	stats.Cutoff = cutoff

	// Estimate prunable events (this is approximate)
	if oldest != nil && oldest.Before(cutoff) {
		// Some events are old enough to prune
		stats.EstimatedPrunable = int64(float64(total) * 0.1) // Very rough estimate
	}

	return stats, nil
}

// RetentionStats contains retention statistics
type RetentionStats struct {
	KeepDays          int
	PruneOnStart      bool
	TotalEvents       int64
	OldestEvent       time.Time
	NewestEvent       time.Time
	Cutoff            time.Time
	EstimatedPrunable int64
}

// PeriodicPruner runs periodic pruning
type PeriodicPruner struct {
	manager  *RetentionManager
	interval time.Duration
	logger   *Logger
	stopChan chan struct{}
}

// NewPeriodicPruner creates a new periodic pruner
func NewPeriodicPruner(manager *RetentionManager, interval time.Duration, logger *Logger) *PeriodicPruner {
	return &PeriodicPruner{
		manager:  manager,
		interval: interval,
		logger:   logger.WithComponent("periodic-pruner"),
		stopChan: make(chan struct{}),
	}
}

// Start begins periodic pruning
func (p *PeriodicPruner) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.logger.Info("periodic pruner started", "interval", p.interval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("periodic pruner stopped")
			return
		case <-p.stopChan:
			p.logger.Info("periodic pruner stopped")
			return
		case <-ticker.C:
			p.logger.Debug("running periodic prune")
			deleted, err := p.manager.PruneOldEvents(ctx)
			if err != nil {
				p.logger.Error("periodic prune failed", "error", err)
			} else {
				p.logger.Info("periodic prune completed", "deleted", deleted)
			}
		}
	}
}

// Stop stops the periodic pruner
func (p *PeriodicPruner) Stop() {
	close(p.stopChan)
}
