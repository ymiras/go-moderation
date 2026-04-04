package engine

import (
	"testing"

	"github.com/ymiras/go-moderation/internal/config"
	"github.com/ymiras/go-moderation/internal/matcher"
	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// mockMatcherForService implements matcher.Matcher for service testing.
type mockMatcherForService struct {
	name     string
	hits     []model.HitRecord
	matchErr error
}

func (m *mockMatcherForService) Name() string { return m.name }

func (m *mockMatcherForService) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	return m.hits, m.matchErr
}

var _ matcher.Matcher = (*mockMatcherForService)(nil)

// mockWordBankForService implements storage.WordBank for service testing.
type mockWordBankForService struct {
	keywords map[string]*model.Keyword
}

func newMockWordBankForService() *mockWordBankForService {
	return &mockWordBankForService{keywords: make(map[string]*model.Keyword)}
}

func (m *mockWordBankForService) Load(path string) error { return nil }
func (m *mockWordBankForService) Contains(word string) (bool, *model.Keyword) {
	if kw, ok := m.keywords[word]; ok {
		return true, kw
	}
	return false, nil
}
func (m *mockWordBankForService) Words() []*model.Keyword {
	result := make([]*model.Keyword, 0, len(m.keywords))
	for _, kw := range m.keywords {
		result = append(result, kw)
	}
	return result
}
func (m *mockWordBankForService) Size() int { return len(m.keywords) }

var _ storage.WordBank = (*mockWordBankForService)(nil)

func TestModerationService_Moderate(t *testing.T) {
	wb := newMockWordBankForService()

	t.Run("returns flagged result with hits", func(t *testing.T) {
		matchers := []matcher.Matcher{
			&mockMatcherForService{
				name: "test",
				hits: []model.HitRecord{{Word: "bad", Severity: model.SeverityHigh}},
			},
		}
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "chain",
				WeightedThreshold: 0.5,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "hello bad world"}
		result, err := svc.Moderate(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Flagged {
			t.Error("expected flagged=true")
		}
		if result.Action != model.ActionBlock {
			t.Errorf("expected action=ActionBlock, got %v", result.Action)
		}
		if len(result.Hits) != 1 {
			t.Errorf("expected 1 hit, got %d", len(result.Hits))
		}
	})

	t.Run("returns pass result when no hits", func(t *testing.T) {
		matchers := []matcher.Matcher{
			&mockMatcherForService{name: "test", hits: nil},
		}
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "chain",
				WeightedThreshold: 0.5,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "hello clean world"}
		result, err := svc.Moderate(ctx)

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

	// Note: Fallback on pipeline error is tested via integration tests
	// because the real pipelines (Chain/Parallel/Weighted) don't return errors
	// on matcher errors - they skip and continue.

	t.Run("tracks latency", func(t *testing.T) {
		matchers := []matcher.Matcher{
			&mockMatcherForService{name: "test", hits: nil},
		}
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "chain",
				WeightedThreshold: 0.5,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "hello world"}
		result, err := svc.Moderate(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// LatencyMs should be recorded (may be 0 for fast operations)
		if result.LatencyMs < 0 {
			t.Errorf("expected non-negative latency, got %f", result.LatencyMs)
		}
	})

	t.Run("returns pass when no hits", func(t *testing.T) {
		matchers := []matcher.Matcher{
			&mockMatcherForService{name: "test", hits: nil},
		}
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "chain",
				WeightedThreshold: 0.5,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "hello world"}
		result, err := svc.Moderate(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Action != model.ActionPass {
			t.Errorf("expected ActionPass, got %v", result.Action)
		}
	})
}

func TestNewService(t *testing.T) {
	wb := newMockWordBankForService()
	matchers := []matcher.Matcher{
		&mockMatcherForService{name: "test", hits: nil},
	}

	t.Run("creates weighted pipeline", func(t *testing.T) {
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "weighted",
				WeightedThreshold: 0.3,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "test"}
		result, _ := svc.Moderate(ctx)
		if result.Action != model.ActionPass {
			t.Errorf("expected ActionPass, got %v", result.Action)
		}
	})

	t.Run("creates parallel pipeline", func(t *testing.T) {
		svc := NewService(&config.Config{
			Moderation: config.ModerationConfig{
				PipelineMode:      "parallel",
				WeightedThreshold: 0.5,
				FallbackAction:    "pass",
			},
		}, wb, matchers)

		ctx := &ModerationContext{Text: "test"}
		result, _ := svc.Moderate(ctx)
		if result.Action != model.ActionPass {
			t.Errorf("expected ActionPass, got %v", result.Action)
		}
	})
}
