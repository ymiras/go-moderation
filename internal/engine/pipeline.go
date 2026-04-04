package engine

import (
	"sync"

	"github.com/ymiras/go-moderation/internal/matcher"
	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// Pipeline defines the interface for executing matchers.
type Pipeline interface {
	Execute(ctx *ModerationContext, wordBank storage.WordBank) (*model.ModerationResult, error)
}

// ChainPipeline executes matchers sequentially, stopping on first hit.
type ChainPipeline struct {
	matchers []matcher.Matcher
}

// NewChainPipeline creates a new ChainPipeline.
func NewChainPipeline(matchers []matcher.Matcher) *ChainPipeline {
	return &ChainPipeline{matchers: matchers}
}

// Execute runs matchers in sequence, stopping at first hit.
func (p *ChainPipeline) Execute(ctx *ModerationContext, wordBank storage.WordBank) (*model.ModerationResult, error) {
	for _, m := range p.matchers {
		hits, err := m.Match(ctx.Text, wordBank)
		if err != nil {
			continue // Skip matcher on error
		}
		if len(hits) > 0 {
			return &model.ModerationResult{
				Flagged: true,
				Action:  model.ActionBlock,
				Hits:    hits,
			}, nil
		}
	}
	return &model.ModerationResult{
		Flagged: false,
		Action:  model.ActionPass,
		Hits:    nil,
	}, nil
}

// ParallelPipeline executes matchers concurrently and merges results.
type ParallelPipeline struct {
	matchers []matcher.Matcher
}

// NewParallelPipeline creates a new ParallelPipeline.
func NewParallelPipeline(matchers []matcher.Matcher) *ParallelPipeline {
	return &ParallelPipeline{matchers: matchers}
}

// Execute runs all matchers concurrently and deduplicates hits.
func (p *ParallelPipeline) Execute(ctx *ModerationContext, wordBank storage.WordBank) (*model.ModerationResult, error) {
	type result struct {
		hits []model.HitRecord
		err  error
	}

	// Run all matchers concurrently
	results := make(chan result, len(p.matchers))
	var wg sync.WaitGroup
	for _, m := range p.matchers {
		wg.Add(1)
		go func(matcher matcher.Matcher) {
			defer wg.Done()
			hits, err := matcher.Match(ctx.Text, wordBank)
			results <- result{hits: hits, err: err}
		}(m)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and deduplicate hits
	seen := make(map[string]bool)
	var allHits []model.HitRecord
	for r := range results {
		if r.err != nil {
			continue
		}
		for _, hit := range r.hits {
			if !seen[hit.Word] {
				seen[hit.Word] = true
				allHits = append(allHits, hit)
			}
		}
	}

	if len(allHits) > 0 {
		return &model.ModerationResult{
			Flagged: true,
			Action:  model.ActionBlock,
			Hits:    allHits,
		}, nil
	}
	return &model.ModerationResult{
		Flagged: false,
		Action:  model.ActionPass,
		Hits:    nil,
	}, nil
}

// WeightedPipeline calculates a weighted score based on hit severities.
type WeightedPipeline struct {
	matchers  []matcher.Matcher
	threshold float64
}

// NewWeightedPipeline creates a new WeightedPipeline.
func NewWeightedPipeline(matchers []matcher.Matcher, threshold float64) *WeightedPipeline {
	return &WeightedPipeline{
		matchers:  matchers,
		threshold: threshold,
	}
}

// severityWeight returns the weight for a given severity.
func severityWeight(s model.Severity) float64 {
	switch s {
	case model.SeverityHigh:
		return 1.0
	case model.SeverityMedium:
		return 0.5
	case model.SeverityLow:
		return 0.2
	default:
		return 0.0
	}
}

// Execute runs all matchers and calculates weighted score.
func (p *WeightedPipeline) Execute(ctx *ModerationContext, wordBank storage.WordBank) (*model.ModerationResult, error) {
	var score float64

	for _, m := range p.matchers {
		hits, err := m.Match(ctx.Text, wordBank)
		if err != nil {
			continue
		}
		for _, hit := range hits {
			score += severityWeight(hit.Severity)
		}
	}

	if score > p.threshold {
		return &model.ModerationResult{
			Flagged: true,
			Action:  model.ActionBlock,
			Hits:    nil, // Weighted mode doesn't return individual hits
		}, nil
	}
	return &model.ModerationResult{
		Flagged: false,
		Action:  model.ActionPass,
		Hits:    nil,
	}, nil
}

// NewPipeline creates a pipeline based on the mode.
func NewPipeline(mode string, matchers []matcher.Matcher, threshold float64) Pipeline {
	switch mode {
	case "parallel":
		return NewParallelPipeline(matchers)
	case "weighted":
		return NewWeightedPipeline(matchers, threshold)
	case "chain":
		fallthrough
	default:
		return NewChainPipeline(matchers)
	}
}
