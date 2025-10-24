package config

import "fmt"

// AdvancedRetention defines sophisticated retention rules
type AdvancedRetention struct {
	Enabled    bool              `yaml:"enabled"`
	Mode       string            `yaml:"mode"` // "rules" or "caps"
	GlobalCaps GlobalCaps        `yaml:"global_caps"`
	Rules      []RetentionRule   `yaml:"rules"`
	Evaluation EvaluationConfig  `yaml:"evaluation"`
}

// GlobalCaps defines hard limits on storage
type GlobalCaps struct {
	MaxTotalEvents   int            `yaml:"max_total_events"`
	MaxStorageMB     int            `yaml:"max_storage_mb"`
	MaxEventsPerKind map[int]int    `yaml:"max_events_per_kind"`
}

// RetentionRule defines a single retention rule
type RetentionRule struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Priority    int                `yaml:"priority"`
	Conditions  RuleConditions     `yaml:"conditions"`
	Action      RetentionAction    `yaml:"action"`
}

// RuleConditions defines the gates for a rule
type RuleConditions struct {
	// Time-based
	CreatedAfter  string `yaml:"created_after"`   // ISO 8601
	CreatedBefore string `yaml:"created_before"`  // ISO 8601
	AgeDaysMax    int    `yaml:"age_days_max"`
	AgeDaysMin    int    `yaml:"age_days_min"`

	// Size-based
	ContentSizeMax int `yaml:"content_size_max"`
	ContentSizeMin int `yaml:"content_size_min"`
	TagsCountMax   int `yaml:"tags_count_max"`

	// Quantity-based
	KindCountMax         map[int]int `yaml:"kind_count_max"`
	AuthorEventCountMax  int         `yaml:"author_event_count_max"`
	AuthorEventCountMin  int         `yaml:"author_event_count_min"`

	// Kind-based
	Kinds        []int  `yaml:"kinds"`
	KindsExclude []int  `yaml:"kinds_exclude"`
	KindCategory string `yaml:"kind_category"` // ephemeral|replaceable|parameterized|regular

	// Social distance
	SocialDistanceMax int      `yaml:"social_distance_max"` // 0=owner, 1=following, 2=FOAF
	SocialDistanceMin int      `yaml:"social_distance_min"`
	AuthorIsOwner     bool     `yaml:"author_is_owner"`
	AuthorIsFollowing bool     `yaml:"author_is_following"`
	AuthorIsMutual    bool     `yaml:"author_is_mutual"`
	AuthorInList      []string `yaml:"author_in_list"`      // npub or hex
	AuthorNotInList   []string `yaml:"author_not_in_list"`

	// Reference-based
	ReferencesOwnerEvents bool     `yaml:"references_owner_events"`
	ReferencesEventIDs    []string `yaml:"references_event_ids"`
	IsRootPost            bool     `yaml:"is_root_post"`
	IsReply               bool     `yaml:"is_reply"`
	HasReplies            bool     `yaml:"has_replies"`
	ReplyCountMin         int      `yaml:"reply_count_min"`
	ReactionCountMin      int      `yaml:"reaction_count_min"`
	ZapSatsMin            int64    `yaml:"zap_sats_min"`

	// Logical operators
	And []RuleConditions `yaml:"and"`
	Or  []RuleConditions `yaml:"or"`
	Not []RuleConditions `yaml:"not"`

	// Catch-all
	All bool `yaml:"all"` // Matches everything
}

// RetentionAction defines what to do with matched events
type RetentionAction struct {
	Retain          bool   `yaml:"retain"`           // Keep forever
	RetainDays      int    `yaml:"retain_days"`      // Keep for N days from created_at
	RetainUntil     string `yaml:"retain_until"`     // Keep until specific date (ISO 8601)
	Delete          bool   `yaml:"delete"`           // Delete on next prune
	DeleteAfterDays int    `yaml:"delete_after_days"` // Grace period before deletion
}

// EvaluationConfig controls when/how rules are evaluated
type EvaluationConfig struct {
	OnIngest           bool `yaml:"on_ingest"`              // Evaluate when event first stored
	ReEvalIntervalHrs  int  `yaml:"re_eval_interval_hours"` // Re-evaluate periodically
	BatchSize          int  `yaml:"batch_size"`             // Process in batches
}

// Validate checks if advanced retention config is valid
func (a *AdvancedRetention) Validate() error {
	if !a.Enabled {
		return nil // Not enabled, no validation needed
	}

	if a.Mode != "rules" && a.Mode != "caps" {
		return fmt.Errorf("invalid advanced.mode: %s (must be 'rules' or 'caps')", a.Mode)
	}

	// Validate rules
	for i, rule := range a.Rules {
		if rule.Name == "" {
			return fmt.Errorf("advanced.rules[%d].name is required", i)
		}
		if rule.Priority < 0 {
			return fmt.Errorf("advanced.rules[%d].priority must be >= 0", i)
		}
		if err := rule.Action.Validate(); err != nil {
			return err
		}
	}

	// Set defaults for evaluation
	if a.Evaluation.OnIngest {
		if a.Evaluation.BatchSize == 0 {
			a.Evaluation.BatchSize = 1000 // Default batch size
		}
	}
	if a.Evaluation.ReEvalIntervalHrs == 0 {
		a.Evaluation.ReEvalIntervalHrs = 168 // Default: weekly
	}

	return nil
}

// Validate checks if retention action is valid
func (a *RetentionAction) Validate() error {
	// Must specify at least one action
	if !a.Retain && a.RetainDays == 0 && a.RetainUntil == "" && !a.Delete {
		return fmt.Errorf("action must specify retain, retain_days, retain_until, or delete")
	}

	// Can't have conflicting actions
	if a.Retain && a.Delete {
		return fmt.Errorf("action cannot have both retain=true and delete=true")
	}

	return nil
}
