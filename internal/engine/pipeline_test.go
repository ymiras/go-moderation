package engine

import (
	"errors"
	"testing"

	"github.com/ymiras/dify-moderation/internal/matcher"
	"github.com/ymiras/dify-moderation/internal/model"
	"github.com/ymiras/dify-moderation/internal/storage"
)

// mockMatcher implements matcher.Matcher for testing.
type mockMatcher struct {
	name     string
	hits     []model.HitRecord
	matchErr error
}

func (m *mockMatcher) Name() string { return m.name }

func (m *mockMatcher) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	return m.hits, m.matchErr
}

var _ matcher.Matcher = (*mockMatcher)(nil)

// mockWordBank implements storage.WordBank for testing.
type mockWordBank struct {
	keywords map[string]*model.Keyword
}

func newMockWordBank() *mockWordBank {
	return &mockWordBank{keywords: make(map[string]*model.Keyword)}
}

func (m *mockWordBank) Load(path string) error { return nil }
func (m *mockWordBank) Contains(word string) (bool, *model.Keyword) {
	if kw, ok := m.keywords[word]; ok {
		return true, kw
	}
	return false, nil
}
func (m *mockWordBank) Words() []*model.Keyword {
	result := make([]*model.Keyword, 0, len(m.keywords))
	for _, kw := range m.keywords {
		result = append(result, kw)
	}
	return result
}
func (m *mockWordBank) Size() int { return len(m.keywords) }

var _ storage.WordBank = (*mockWordBank)(nil)

func toMatchers(mm []mockMatcher) []matcher.Matcher {
	result := make([]matcher.Matcher, len(mm))
	for i := range mm {
		result[i] = &mm[i]
	}
	return result
}

func TestChainPipeline_Execute(t *testing.T) {
	wb := newMockWordBank()
	ctx := &ModerationContext{Text: "hello world"}

	t.Run("stops on first hit", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: nil},
			{name: "m2", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
			{name: "m3", hits: []model.HitRecord{{Word: "worse", Severity: model.SeverityHigh}}},
		})
		pipeline := NewChainPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true")
		}
		if len(result.Hits) != 1 || result.Hits[0].Word != "bad" {
			t.Errorf("expected hits=[bad], got %v", result.Hits)
		}
	})

	t.Run("passes when no hits", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: nil},
			{name: "m2", hits: nil},
		})
		pipeline := NewChainPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Flagged {
			t.Error("expected flagged=false")
		}
		if result.Action != model.ActionPass {
			t.Errorf("expected action=ActionPass, got %v", result.Action)
		}
	})

	t.Run("skips matcher on error", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", matchErr: errors.New("matcher error")},
			{name: "m2", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
		})
		pipeline := NewChainPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true")
		}
	})
}

func TestParallelPipeline_Execute(t *testing.T) {
	wb := newMockWordBank()
	ctx := &ModerationContext{Text: "hello world"}

	t.Run("deduplicates hits by word", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
			{name: "m2", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityMedium}, {Word: "worse", Severity: model.SeverityHigh}}},
		})
		pipeline := NewParallelPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true")
		}
		if len(result.Hits) != 2 {
			t.Errorf("expected 2 hits (deduplicated), got %d: %v", len(result.Hits), result.Hits)
		}
	})

	t.Run("passes when no hits", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: nil},
			{name: "m2", hits: nil},
		})
		pipeline := NewParallelPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Flagged {
			t.Error("expected flagged=false")
		}
		if result.Action != model.ActionPass {
			t.Errorf("expected action=ActionPass, got %v", result.Action)
		}
	})

	t.Run("skips matchers with errors", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", matchErr: errors.New("matcher error")},
			{name: "m2", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
		})
		pipeline := NewParallelPipeline(matchers)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true")
		}
	})
}

func TestWeightedPipeline_Execute(t *testing.T) {
	wb := newMockWordBank()
	ctx := &ModerationContext{Text: "hello world"}

	t.Run("blocks when score exceeds threshold", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{
				{Word: "bad1", Severity: model.SeverityHigh},   // 1.0
				{Word: "bad2", Severity: model.SeverityMedium}, // 0.5
			}},
		})
		pipeline := NewWeightedPipeline(matchers, 0.5)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true (score=1.5 > threshold=0.5)")
		}
		if result.Action != model.ActionBlock {
			t.Errorf("expected action=ActionBlock, got %v", result.Action)
		}
	})

	t.Run("passes when score equals threshold", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{
				{Word: "bad", Severity: model.SeverityMedium}, // 0.5
			}},
		})
		pipeline := NewWeightedPipeline(matchers, 0.5)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Flagged {
			t.Error("expected flagged=false (score=0.5 == threshold=0.5, not >)")
		}
	})

	t.Run("passes when score below threshold", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{
				{Word: "bad", Severity: model.SeverityLow}, // 0.2
			}},
		})
		pipeline := NewWeightedPipeline(matchers, 0.5)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Flagged {
			t.Error("expected flagged=false (score=0.2 < threshold=0.5)")
		}
	})

	t.Run("accumulates score across matchers", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{{Word: "bad1", Severity: model.SeverityMedium}}}, // 0.5
			{name: "m2", hits: []model.HitRecord{{Word: "bad2", Severity: model.SeverityLow}}},    // 0.2
		})
		pipeline := NewWeightedPipeline(matchers, 0.6)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true (score=0.7 > threshold=0.6)")
		}
	})

	t.Run("weighted mode returns nil hits", func(t *testing.T) {
		matchers := toMatchers([]mockMatcher{
			{name: "m1", hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}}},
		})
		pipeline := NewWeightedPipeline(matchers, 0.1)

		result, err := pipeline.Execute(ctx, wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Hits != nil {
			t.Errorf("expected nil hits in weighted mode, got %v", result.Hits)
		}
	})
}

func TestNewPipeline(t *testing.T) {
	matchers := toMatchers([]mockMatcher{{name: "m1", hits: nil}})

	t.Run("chain mode", func(t *testing.T) {
		p := NewPipeline("chain", matchers, 0.5)
		if _, ok := p.(*ChainPipeline); !ok {
			t.Error("expected ChainPipeline")
		}
	})

	t.Run("parallel mode", func(t *testing.T) {
		p := NewPipeline("parallel", matchers, 0.5)
		if _, ok := p.(*ParallelPipeline); !ok {
			t.Error("expected ParallelPipeline")
		}
	})

	t.Run("weighted mode", func(t *testing.T) {
		p := NewPipeline("weighted", matchers, 0.5)
		if _, ok := p.(*WeightedPipeline); !ok {
			t.Error("expected WeightedPipeline")
		}
	})

	t.Run("unknown mode defaults to chain", func(t *testing.T) {
		p := NewPipeline("unknown", matchers, 0.5)
		if _, ok := p.(*ChainPipeline); !ok {
			t.Error("expected ChainPipeline as default")
		}
	})
}
