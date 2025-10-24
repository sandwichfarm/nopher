package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAdvancedRetentionParsing(t *testing.T) {
	yamlConfig := `
sync:
  retention:
    keep_days: 90
    prune_on_start: false
    advanced:
      enabled: true
      mode: "rules"
      global_caps:
        max_total_events: 10000
        max_storage_mb: 100
      evaluation:
        on_ingest: true
        re_eval_interval_hours: 168
        batch_size: 1000
      rules:
        - name: "protect_owner"
          description: "Never delete owner's content"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true
        - name: "default"
          description: "Default retention"
          priority: 100
          conditions:
            all: true
          action:
            retain_days: 90
`

	type testConfig struct {
		Sync struct {
			Retention Retention `yaml:"retention"`
		} `yaml:"sync"`
	}

	var cfg testConfig
	if err := yaml.Unmarshal([]byte(yamlConfig), &cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Check basic fields
	if cfg.Sync.Retention.KeepDays != 90 {
		t.Errorf("Expected keep_days=90, got %d", cfg.Sync.Retention.KeepDays)
	}

	// Check advanced retention
	if cfg.Sync.Retention.Advanced == nil {
		t.Fatal("Advanced retention is nil")
	}

	adv := cfg.Sync.Retention.Advanced
	if !adv.Enabled {
		t.Error("Advanced retention should be enabled")
	}

	if adv.Mode != "rules" {
		t.Errorf("Expected mode='rules', got '%s'", adv.Mode)
	}

	if adv.GlobalCaps.MaxTotalEvents != 10000 {
		t.Errorf("Expected max_total_events=10000, got %d", adv.GlobalCaps.MaxTotalEvents)
	}

	if len(adv.Rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(adv.Rules))
	}

	// Check first rule
	rule1 := adv.Rules[0]
	if rule1.Name != "protect_owner" {
		t.Errorf("Expected rule name 'protect_owner', got '%s'", rule1.Name)
	}
	if rule1.Priority != 1000 {
		t.Errorf("Expected priority=1000, got %d", rule1.Priority)
	}
	if !rule1.Conditions.AuthorIsOwner {
		t.Error("Expected author_is_owner=true")
	}
	if !rule1.Action.Retain {
		t.Error("Expected action.retain=true")
	}

	// Check second rule
	rule2 := adv.Rules[1]
	if rule2.Name != "default" {
		t.Errorf("Expected rule name 'default', got '%s'", rule2.Name)
	}
	if !rule2.Conditions.All {
		t.Error("Expected conditions.all=true")
	}
	if rule2.Action.RetainDays != 90 {
		t.Errorf("Expected retain_days=90, got %d", rule2.Action.RetainDays)
	}

	t.Log("✅ Advanced retention parsing test passed!")
}

func TestAdvancedRetentionValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *AdvancedRetention
		wantErr bool
	}{
		{
			name: "valid config",
			config: &AdvancedRetention{
				Enabled: true,
				Mode:    "rules",
				Rules: []RetentionRule{
					{
						Name:     "test",
						Priority: 100,
						Action:   RetentionAction{Retain: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: &AdvancedRetention{
				Enabled: true,
				Mode:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "missing rule name",
			config: &AdvancedRetention{
				Enabled: true,
				Mode:    "rules",
				Rules: []RetentionRule{
					{
						Name:     "",
						Priority: 100,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative priority",
			config: &AdvancedRetention{
				Enabled: true,
				Mode:    "rules",
				Rules: []RetentionRule{
					{
						Name:     "test",
						Priority: -1,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
