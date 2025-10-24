package sections

import (
	"fmt"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// FilterBuilder helps construct complex filters for sections
type FilterBuilder struct {
	filter nostr.Filter
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filter: nostr.Filter{},
	}
}

// Kinds adds kind filters
func (fb *FilterBuilder) Kinds(kinds ...int) *FilterBuilder {
	fb.filter.Kinds = append(fb.filter.Kinds, kinds...)
	return fb
}

// Authors adds author filters
func (fb *FilterBuilder) Authors(authors ...string) *FilterBuilder {
	fb.filter.Authors = append(fb.filter.Authors, authors...)
	return fb
}

// IDs adds ID filters
func (fb *FilterBuilder) IDs(ids ...string) *FilterBuilder {
	fb.filter.IDs = append(fb.filter.IDs, ids...)
	return fb
}

// Tag adds a tag filter
func (fb *FilterBuilder) Tag(key string, values ...string) *FilterBuilder {
	if fb.filter.Tags == nil {
		fb.filter.Tags = make(nostr.TagMap)
	}
	fb.filter.Tags[key] = append(fb.filter.Tags[key], values...)
	return fb
}

// Since sets the since timestamp
func (fb *FilterBuilder) Since(t time.Time) *FilterBuilder {
	since := nostr.Timestamp(t.Unix())
	fb.filter.Since = &since
	return fb
}

// Until sets the until timestamp
func (fb *FilterBuilder) Until(t time.Time) *FilterBuilder {
	until := nostr.Timestamp(t.Unix())
	fb.filter.Until = &until
	return fb
}

// Limit sets the result limit
func (fb *FilterBuilder) Limit(limit int) *FilterBuilder {
	fb.filter.Limit = limit
	return fb
}

// Build returns the constructed filter
func (fb *FilterBuilder) Build() nostr.Filter {
	return fb.filter
}

// ScopeFilterBuilder builds filters based on social graph scope
type ScopeFilterBuilder struct {
	ownerPubkey string
	scope       Scope
	depth       int
}

// NewScopeFilterBuilder creates a new scope filter builder
func NewScopeFilterBuilder(ownerPubkey string, scope Scope, depth int) *ScopeFilterBuilder {
	return &ScopeFilterBuilder{
		ownerPubkey: ownerPubkey,
		scope:       scope,
		depth:       depth,
	}
}

// BuildAuthors returns the list of authors based on scope
// This is a simplified version - full implementation would query the graph
func (sfb *ScopeFilterBuilder) BuildAuthors() ([]string, error) {
	switch sfb.scope {
	case ScopeSelf:
		return []string{sfb.ownerPubkey}, nil
	case ScopeFollowing:
		// Would query graph for following list
		return []string{sfb.ownerPubkey}, nil
	case ScopeMutual:
		// Would query graph for mutual follows
		return []string{sfb.ownerPubkey}, nil
	case ScopeFoaf:
		// Would query graph for friends-of-friends
		return []string{sfb.ownerPubkey}, nil
	case ScopeAll:
		// No author filter - return all
		return []string{}, nil
	default:
		return nil, fmt.Errorf("unknown scope: %s", sfb.scope)
	}
}

// TimeRangeFilter creates a filter for a specific time range
type TimeRangeFilter struct {
	Start time.Time
	End   time.Time
}

// NewTimeRangeFilter creates a new time range filter
func NewTimeRangeFilter(start, end time.Time) *TimeRangeFilter {
	return &TimeRangeFilter{
		Start: start,
		End:   end,
	}
}

// Apply applies the time range to a filter builder
func (trf *TimeRangeFilter) Apply(fb *FilterBuilder) *FilterBuilder {
	if !trf.Start.IsZero() {
		fb.Since(trf.Start)
	}
	if !trf.End.IsZero() {
		fb.Until(trf.End)
	}
	return fb
}

// Today returns a time range for today
func Today() *TimeRangeFilter {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return NewTimeRangeFilter(start, end)
}

// Yesterday returns a time range for yesterday
func Yesterday() *TimeRangeFilter {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return NewTimeRangeFilter(start, end)
}

// ThisWeek returns a time range for this week
func ThisWeek() *TimeRangeFilter {
	now := time.Now()
	weekday := int(now.Weekday())
	start := now.AddDate(0, 0, -weekday)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.AddDate(0, 0, 7)
	return NewTimeRangeFilter(start, end)
}

// ThisMonth returns a time range for this month
func ThisMonth() *TimeRangeFilter {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return NewTimeRangeFilter(start, end)
}

// ThisYear returns a time range for this year
func ThisYear() *TimeRangeFilter {
	now := time.Now()
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(1, 0, 0)
	return NewTimeRangeFilter(start, end)
}

// LastNDays returns a time range for the last N days
func LastNDays(n int) *TimeRangeFilter {
	end := time.Now()
	start := end.AddDate(0, 0, -n)
	return NewTimeRangeFilter(start, end)
}

// LastNHours returns a time range for the last N hours
func LastNHours(n int) *TimeRangeFilter {
	end := time.Now()
	start := end.Add(time.Duration(-n) * time.Hour)
	return NewTimeRangeFilter(start, end)
}

// KindFilter creates common kind-based filters
type KindFilter int

const (
	KindFilterNotes      KindFilter = 1
	KindFilterArticles   KindFilter = 30023
	KindFilterReactions  KindFilter = 7
	KindFilterZaps       KindFilter = 9735
	KindFilterReposts    KindFilter = 6
	KindFilterProfiles   KindFilter = 0
	KindFilterContacts   KindFilter = 3
	KindFilterRelayLists KindFilter = 10002
)

// Apply applies the kind filter to a filter builder
func (kf KindFilter) Apply(fb *FilterBuilder) *FilterBuilder {
	return fb.Kinds(int(kf))
}

// CombineFilters combines multiple filters with OR logic
func CombineFilters(filters ...nostr.Filter) []nostr.Filter {
	return filters
}

// InteractionFilter creates a filter for interactions with a specific event
func InteractionFilter(eventID string) *FilterBuilder {
	return NewFilterBuilder().
		Kinds(7, 9735, 1). // Reactions, zaps, replies
		Tag("e", eventID)
}

// ThreadFilter creates a filter for a thread
func ThreadFilter(rootEventID string) *FilterBuilder {
	return NewFilterBuilder().
		Kinds(1).
		Tag("e", rootEventID)
}

// MentionFilter creates a filter for mentions of a pubkey
func MentionFilter(pubkey string) *FilterBuilder {
	return NewFilterBuilder().
		Kinds(1).
		Tag("p", pubkey)
}

// ReplyFilter creates a filter for replies to a pubkey
func ReplyFilter(pubkey string) *FilterBuilder {
	return NewFilterBuilder().
		Kinds(1).
		Tag("p", pubkey)
}
