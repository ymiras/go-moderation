package engine

import (
	"testing"

	"github.com/ymiras/go-moderation/internal/matcher"
	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// testMatcher implements matcher.Matcher for engine testing.
type testMatcher struct {
	name string
	hits []model.HitRecord
}

func (m *testMatcher) Name() string { return m.name }

func (m *testMatcher) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	return m.hits, nil
}

var _ matcher.Matcher = (*testMatcher)(nil)

func TestRuleEngine_PriorityConflictResolution(t *testing.T) {
	store := NewMemoryRuleStore()
	store.AddRules([]Rule{
		{
			ID:       "low_priority_block",
			Matcher:  "ac",
			Action:   model.ActionBlock,
			Priority: 1,
		},
		{
			ID:       "high_priority_mask",
			Matcher:  "ac",
			Action:   model.ActionMask,
			Priority: 10,
		},
	})

	matchers := []matcher.Matcher{
		&testMatcher{name: "ac", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
	}

	engine := NewRuleEngine(store, matchers)

	ctx := &ModerationContext{Text: "hello bad world"}
	decision := engine.Evaluate(ctx)

	// High priority should win
	if decision.RuleID != "high_priority_mask" {
		t.Errorf("expected rule_id=high_priority_mask, got %s", decision.RuleID)
	}
	if decision.Action != model.ActionMask {
		t.Errorf("expected action=mask, got %s", decision.Action)
	}
}

func TestRuleEngine_SamePriorityDifferentAction(t *testing.T) {
	store := NewMemoryRuleStore()
	store.AddRules([]Rule{
		{
			ID:       "pass_rule",
			Matcher:  "ac",
			Action:   model.ActionPass,
			Priority: 5,
		},
		{
			ID:       "block_rule",
			Matcher:  "ac",
			Action:   model.ActionBlock,
			Priority: 5,
		},
	})

	matchers := []matcher.Matcher{
		&testMatcher{name: "ac", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
	}

	engine := NewRuleEngine(store, matchers)

	ctx := &ModerationContext{Text: "hello bad world"}
	decision := engine.Evaluate(ctx)

	// Same priority, should take more severe action (block > pass)
	if decision.RuleID != "block_rule" {
		t.Errorf("expected rule_id=block_rule, got %s", decision.RuleID)
	}
	if decision.Action != model.ActionBlock {
		t.Errorf("expected action=block, got %s", decision.Action)
	}
}
