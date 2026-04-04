package matcher

import (
	"testing"

	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// mockWordBank implements storage.WordBank for testing.
type mockWordBank struct {
	keywords []*model.Keyword
}

func (m *mockWordBank) Load(path string) error                      { return nil }
func (m *mockWordBank) Contains(word string) (bool, *model.Keyword) { return false, nil }
func (m *mockWordBank) Words() []*model.Keyword                     { return m.keywords }
func (m *mockWordBank) Size() int                                   { return len(m.keywords) }

var _ storage.WordBank = (*mockWordBank)(nil)

func TestACMatcher_Match(t *testing.T) {
	t.Run("basic match", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("hello bad world", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 1 {
			t.Fatalf("expected 1 hit, got %d", len(hits))
		}
		if hits[0].Word != "bad" {
			t.Errorf("expected word 'bad', got '%s'", hits[0].Word)
		}
	})

	t.Run("no match", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("hello world", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 0 {
			t.Errorf("expected 0 hits, got %d", len(hits))
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
				{Word: "evil", Type: "profanity", Severity: model.SeverityMedium},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("this is bad and evil", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 2 {
			t.Errorf("expected 2 hits, got %d", len(hits))
		}
	})

	t.Run("multiple same keyword", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("bad bad bad", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 3 {
			t.Errorf("expected 3 hits, got %d", len(hits))
		}
	})

	t.Run("overlapping patterns", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "ab", Type: "test", Severity: model.SeverityLow},
				{Word: "bc", Type: "test", Severity: model.SeverityLow},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("abc", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Both "ab" and "bc" should match at positions 0 and 1
		if len(hits) != 2 {
			t.Errorf("expected 2 hits, got %d: %v", len(hits), hits)
		}
	})

	t.Run("hit record fields", func(t *testing.T) {
		wb := &mockWordBank{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &ACMatcher{}
		hits, err := m.Match("hel lo bad wo rld", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 1 {
			t.Fatalf("expected 1 hit, got %d", len(hits))
		}
		hit := hits[0]
		if hit.Index != 7 {
			t.Errorf("expected Index 7, got %d", hit.Index)
		}
		if hit.Length != 3 {
			t.Errorf("expected Length 3, got %d", hit.Length)
		}
		if hit.Type != "profanity" {
			t.Errorf("expected Type 'profanity', got '%s'", hit.Type)
		}
		if hit.Severity != model.SeverityHigh {
			t.Errorf("expected Severity High, got %v", hit.Severity)
		}
	})
}

func TestACMatcher_Name(t *testing.T) {
	m := &ACMatcher{}
	if m.Name() != "ac-automaton" {
		t.Errorf("expected 'ac-automaton', got '%s'", m.Name())
	}
}

func TestACMatcher_LazyBuild(t *testing.T) {
	wb := &mockWordBank{
		keywords: []*model.Keyword{
			{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
		},
	}

	m := &ACMatcher{}

	// Should not build until first Match
	m.mu.RLock()
	if m.built {
		t.Error("expected not built before Match")
	}
	m.mu.RUnlock()

	_, err := m.Match("hello bad world", wb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be built after Match
	m.mu.RLock()
	if !m.built {
		t.Error("expected built after Match")
	}
	m.mu.RUnlock()
}
