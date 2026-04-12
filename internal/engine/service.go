package engine

import (
	"time"

	"github.com/ymiras/go-moderation/internal/config"
	"github.com/ymiras/go-moderation/internal/matcher"
	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// ModerationService is the entry point for content moderation.
type ModerationService struct {
	wordBank   storage.WordBank
	ruleEngine *RuleEngine
	fallback   model.ActionType
}

// NewService creates a new ModerationService.
func NewService(cfg *config.Config, wordBank storage.WordBank, matchers []matcher.Matcher) *ModerationService {
	fallback := model.ActionPass
	if cfg.Moderation.FallbackAction == "block" {
		fallback = model.ActionBlock
	}

	// Create rule engine
	ruleEngine := NewRuleEngine(nil, matchers)

	return &ModerationService{
		wordBank:   wordBank,
		ruleEngine: ruleEngine,
		fallback:   fallback,
	}
}

// SetRuleStore sets the rule store for the moderation service.
func (s *ModerationService) SetRuleStore(store RuleStore) {
	s.ruleEngine.store = store
}

// Moderate performs content moderation on the given context.
func (s *ModerationService) Moderate(ctx *ModerationContext) (*model.ModerationResult, error) {
	start := time.Now()

	// Use rule engine
	decision := s.ruleEngine.Evaluate(ctx)

	if decision.Action == "" {
		decision.Action = s.fallback
	}

	return &model.ModerationResult{
		Flagged:   decision.Action != model.ActionPass,
		Action:    decision.Action,
		Hits:      decision.Hits,
		LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}, nil
}
