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
	wordBank storage.WordBank
	pipeline Pipeline
	fallback model.ActionType
}

// NewService creates a new ModerationService.
func NewService(cfg *config.Config, wordBank storage.WordBank, matchers []matcher.Matcher) *ModerationService {
	fallback := model.ActionPass
	if cfg.Moderation.FallbackAction == "block" {
		fallback = model.ActionBlock
	}

	pipeline := NewPipeline(
		cfg.Moderation.PipelineMode,
		matchers,
		cfg.Moderation.WeightedThreshold,
	)

	return &ModerationService{
		wordBank: wordBank,
		pipeline: pipeline,
		fallback: fallback,
	}
}

// Moderate performs content moderation on the given context.
func (s *ModerationService) Moderate(ctx *ModerationContext) (*model.ModerationResult, error) {
	start := time.Now()

	result, err := s.pipeline.Execute(ctx, s.wordBank)
	if err != nil {
		// Fallback on error
		return &model.ModerationResult{
			Flagged:   false,
			Action:    s.fallback,
			Hits:      nil,
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}, nil
	}

	result.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
	return result, nil
}
