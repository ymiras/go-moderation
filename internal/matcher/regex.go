package matcher

import (
	"regexp"
	"strings"
	"sync"

	"github.com/ymiras/dify-moderation/internal/model"
	"github.com/ymiras/dify-moderation/internal/storage"
)

// RegexConfig is the configuration for the regex matcher.
// Currently empty, reserved for future options.
type RegexConfig struct{}

// RegexMatcher implements the Matcher interface using regex pattern matching.
type RegexMatcher struct {
	cfg    *RegexConfig
	mu     sync.RWMutex
	cache  map[string]*regexp.Regexp
	loaded bool
}

// NewRegex New creates a new regex matcher.
func NewRegex(cfg *RegexConfig) (Matcher, error) {
	return &RegexMatcher{
		cfg:   cfg,
		cache: make(map[string]*regexp.Regexp),
	}, nil
}

// Name returns the matcher name.
func (m *RegexMatcher) Name() string {
	return "regex-matcher"
}

// Match searches for all regex patterns in the text.
func (m *RegexMatcher) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	m.mu.RLock()
	loaded := m.loaded
	m.mu.RUnlock()

	if !loaded {
		if err := m.ensureCache(wordBank); err != nil {
			return nil, err
		}
	}

	// For now, iterate through all keywords and treat them as literal strings
	// If a keyword contains regex metacharacters, it will be treated as a regex pattern
	keywords := wordBank.Words()
	var hits []model.HitRecord

	for _, kw := range keywords {
		pattern, err := m.getPattern(kw.Word)
		if err != nil {
			continue // Skip invalid patterns
		}

		loc := pattern.FindStringIndex(text)
		if loc != nil {
			hits = append(hits, model.HitRecord{
				Word:     kw.Word,
				Type:     kw.Type,
				Severity: kw.Severity,
				Index:    loc[0],
				Length:   loc[1] - loc[0],
			})
		}
	}

	return hits, nil
}

// getPattern returns a compiled regex for the given pattern string.
func (m *RegexMatcher) getPattern(s string) (*regexp.Regexp, error) {
	// Check cache first
	m.mu.RLock()
	re, ok := m.cache[s]
	m.mu.RUnlock()

	if ok {
		return re, nil
	}

	// Compile and cache
	// If the pattern is a simple string (no regex metacharacters),
	// we could use exact match, but for simplicity we compile as regex
	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.cache[s] = re
	m.mu.Unlock()

	return re, nil
}

// ensureCache pre-loads all keyword patterns into the cache.
func (m *RegexMatcher) ensureCache(wordBank storage.WordBank) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		return nil
	}

	if m.cache == nil {
		m.cache = make(map[string]*regexp.Regexp)
	}

	keywords := wordBank.Words()
	for _, kw := range keywords {
		// Use the word as-is as the regex pattern
		re, err := regexp.Compile(kw.Word)
		if err != nil {
			// Skip invalid patterns
			continue
		}
		m.cache[kw.Word] = re
	}

	m.loaded = true
	return nil
}

// LiteralMatch returns a HitRecord if the exact word is found in text.
func LiteralMatch(text string, word string, kw *model.Keyword) *model.HitRecord {
	idx := strings.Index(text, word)
	if idx == -1 {
		return nil
	}
	return &model.HitRecord{
		Word:     kw.Word,
		Type:     kw.Type,
		Severity: kw.Severity,
		Index:    idx,
		Length:   len(word),
	}
}
