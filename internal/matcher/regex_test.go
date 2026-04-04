package matcher

import (
	"testing"

	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// mockWordBankForRegex implements storage.WordBank for regex testing.
type mockWordBankForRegex struct {
	keywords []*model.Keyword
}

func (m *mockWordBankForRegex) Load(path string) error                      { return nil }
func (m *mockWordBankForRegex) Contains(word string) (bool, *model.Keyword) { return false, nil }
func (m *mockWordBankForRegex) Words() []*model.Keyword                     { return m.keywords }
func (m *mockWordBankForRegex) Size() int                                   { return len(m.keywords) }

var _ storage.WordBank = (*mockWordBankForRegex)(nil)

func TestRegexMatcher_Match(t *testing.T) {
	t.Run("simple match", func(t *testing.T) {
		wb := &mockWordBankForRegex{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &RegexMatcher{}
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
		wb := &mockWordBankForRegex{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &RegexMatcher{}
		hits, err := m.Match("hello world", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 0 {
			t.Errorf("expected 0 hits, got %d", len(hits))
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		wb := &mockWordBankForRegex{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
				{Word: "evil", Type: "profanity", Severity: model.SeverityMedium},
			},
		}
		m := &RegexMatcher{}
		hits, err := m.Match("this is bad and evil", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 2 {
			t.Errorf("expected 2 hits, got %d", len(hits))
		}
	})

	t.Run("regex metacharacters", func(t *testing.T) {
		wb := &mockWordBankForRegex{
			keywords: []*model.Keyword{
				{Word: "\\bbad\\b", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &RegexMatcher{}
		hits, err := m.Match("hello bad world", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// \b matches word boundary, so "bad" surrounded by spaces should match
		if len(hits) != 1 {
			t.Errorf("expected 1 hit with word boundary regex, got %d", len(hits))
		}
	})

	t.Run("hit record fields", func(t *testing.T) {
		wb := &mockWordBankForRegex{
			keywords: []*model.Keyword{
				{Word: "bad", Type: "profanity", Severity: model.SeverityHigh},
			},
		}
		m := &RegexMatcher{}
		hits, err := m.Match("hello bad world", wb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 1 {
			t.Fatalf("expected 1 hit, got %d", len(hits))
		}
		hit := hits[0]
		if hit.Index != 6 {
			t.Errorf("expected Index 6, got %d", hit.Index)
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

func TestRegexMatcher_Name(t *testing.T) {
	m := &RegexMatcher{}
	if m.Name() != "regex-matcher" {
		t.Errorf("expected 'regex-matcher', got '%s'", m.Name())
	}
}

func TestLiteralMatch(t *testing.T) {
	kw := &model.Keyword{Word: "bad", Type: "test", Severity: model.SeverityMedium}

	t.Run("found", func(t *testing.T) {
		hit := LiteralMatch("hello bad world", "bad", kw)
		if hit == nil {
			t.Fatal("expected hit, got nil")
		}
		if hit.Index != 6 {
			t.Errorf("expected Index 6, got %d", hit.Index)
		}
	})

	t.Run("not found", func(t *testing.T) {
		hit := LiteralMatch("hello world", "bad", kw)
		if hit != nil {
			t.Errorf("expected nil, got %v", hit)
		}
	})
}
