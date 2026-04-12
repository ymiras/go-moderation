package engine

import (
	"sort"
	"time"

	"github.com/ymiras/go-moderation/internal/matcher"
	"github.com/ymiras/go-moderation/internal/model"
)

// RuleEngine evaluates text against rules.
type RuleEngine struct {
	store    RuleStore
	matchers []matcher.Matcher
}

// NewRuleEngine creates a new RuleEngine.
func NewRuleEngine(store RuleStore, matchers []matcher.Matcher) *RuleEngine {
	return &RuleEngine{
		store:    store,
		matchers: matchers,
	}
}

// Evaluate evaluates the text against the rules.
func (e *RuleEngine) Evaluate(ctx *ModerationContext) *Decision {
	start := time.Now()

	// If store is nil, just run all matchers and return pass (backwards compatible)
	if e.store == nil {
		return e.evaluateWithMatchers(ctx)
	}

	// Get rules from store
	rules := e.store.GetRules()
	if len(rules) == 0 {
		// No rules found, return pass
		return &Decision{
			Action:    model.ActionPass,
			Hits:      nil,
			RuleID:    "",
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}
	}

	// Sort rules by priority (descending)
	sortedRules := make([]Rule, len(rules))
	copy(sortedRules, rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority > sortedRules[j].Priority
	})

	// Execute all matchers and collect results
	matcherHits := make(map[string][]model.HitRecord)
	for _, m := range e.matchers {
		hits, err := m.Match(ctx.Text, nil)
		if err != nil {
			// TODO: handle error (will be done in error-handling-priority change)
			continue
		}
		matcherHits[m.Name()] = hits
	}

	// Merge all hits
	var allHits []model.HitRecord
	for _, hits := range matcherHits {
		allHits = append(allHits, hits...)
	}

	// Evaluate rules in priority order
	var selected *Rule
	for i := range sortedRules {
		rule := &sortedRules[i]
		if e.evaluateRule(rule, matcherHits) {
			if selected == nil {
				selected = rule
			} else if rule.Priority > selected.Priority {
				selected = rule
			} else if rule.Priority == selected.Priority {
				// Same priority, take the more severe action
				if actionPriority(rule.Action) > actionPriority(selected.Action) {
					selected = rule
				}
			}
		}
	}

	if selected == nil {
		return &Decision{
			Action:    model.ActionPass,
			Hits:      nil,
			RuleID:    "",
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}
	}

	return &Decision{
		Action:    selected.Action,
		Hits:      allHits,
		RuleID:    selected.ID,
		LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
}

// evaluateRule checks if a rule matches (simplified - always returns true for now).
// TODO: Implement condition evaluation (condition-system change).
func (e *RuleEngine) evaluateRule(rule *Rule, matcherHits map[string][]model.HitRecord) bool {
	// For now, just check if the matcher has any hits
	hits, ok := matcherHits[rule.Matcher]
	if !ok || len(hits) == 0 {
		return false
	}
	return true
}

// evaluateWithMatchers runs all matchers and returns results (for backwards compatibility when store is nil).
func (e *RuleEngine) evaluateWithMatchers(ctx *ModerationContext) *Decision {
	start := time.Now()

	// Execute all matchers and collect results
	matcherHits := make(map[string][]model.HitRecord)
	for _, m := range e.matchers {
		hits, err := m.Match(ctx.Text, nil)
		if err != nil {
			continue
		}
		matcherHits[m.Name()] = hits
	}

	// Merge all hits
	var allHits []model.HitRecord
	for _, hits := range matcherHits {
		allHits = append(allHits, hits...)
	}

	// If there are hits, return blocked (backwards compatible behavior)
	if len(allHits) > 0 {
		return &Decision{
			Action:    model.ActionBlock,
			Hits:      allHits,
			RuleID:    "",
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}
	}

	return &Decision{
		Action:    model.ActionPass,
		Hits:      nil,
		RuleID:    "",
		LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
}

// actionPriority returns the priority of an action (for conflict resolution).
func actionPriority(action model.ActionType) int {
	switch action {
	case model.ActionBlock:
		return 3
	case model.ActionMask:
		return 2
	default:
		return 1
	}
}
