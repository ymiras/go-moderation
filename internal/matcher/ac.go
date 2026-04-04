package matcher

import (
	"fmt"
	"sync"

	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// ACConfig is the configuration for the AC automaton matcher.
// Currently empty, reserved for future options (e.g., case sensitivity).
type ACConfig struct{}

// ACMatcher implements the Matcher interface using Aho-Corasick algorithm.
type ACMatcher struct {
	cfg   *ACConfig
	built bool
	mu    sync.RWMutex
	root  *acNode
}

type acNode struct {
	children map[rune]*acNode
	fail     *acNode
	output   bool
	keyword  *model.Keyword
}

func newACNode() *acNode {
	return &acNode{
		children: make(map[rune]*acNode),
	}
}

// NewAC New creates a new AC automaton matcher.
func NewAC(cfg *ACConfig) (Matcher, error) {
	return &ACMatcher{cfg: cfg}, nil
}

// Name returns the matcher name.
func (m *ACMatcher) Name() string {
	return "ac-automaton"
}

// Match searches for all keywords in the text.
func (m *ACMatcher) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	if err := m.ensureBuilt(wordBank); err != nil {
		return nil, err
	}

	var hits []model.HitRecord
	node := m.root

	for i, r := range text {
		// Follow fail links until we find a matching child or reach root
		for node.children[r] == nil && node != m.root {
			node = node.fail
		}

		if child, ok := node.children[r]; ok {
			node = child
		}

		// Check for output at current node
		if node.output {
			kw := node.keyword
			hits = append(hits, model.HitRecord{
				Word:     kw.Word,
				Type:     kw.Type,
				Severity: kw.Severity,
				Index:    i - len(kw.Word) + 1,
				Length:   len(kw.Word),
			})
		}

		// Also check fail chain for outputs (keywords that are suffixes of longer matches)
		for fp := node.fail; fp != nil && fp != m.root; fp = fp.fail {
			if fp.output {
				kw := fp.keyword
				hits = append(hits, model.HitRecord{
					Word:     kw.Word,
					Type:     kw.Type,
					Severity: kw.Severity,
					Index:    i - len(kw.Word) + 1,
					Length:   len(kw.Word),
				})
			}
		}
	}

	return hits, nil
}

// ensureBuilt builds the automaton if not already built.
func (m *ACMatcher) ensureBuilt(wordBank storage.WordBank) error {
	m.mu.RLock()
	if m.built {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if m.built {
		return nil
	}

	keywords := wordBank.Words()

	m.root = newACNode()
	for _, kw := range keywords {
		m.insert(kw)
	}

	if err := m.buildFailureLinks(); err != nil {
		return fmt.Errorf("failed to build failure links: %w", err)
	}

	m.built = true
	return nil
}

// insert adds a keyword to the trie.
func (m *ACMatcher) insert(kw *model.Keyword) {
	node := m.root
	for _, r := range kw.Word {
		if node.children[r] == nil {
			node.children[r] = newACNode()
		}
		node = node.children[r]
	}
	node.output = true
	node.keyword = kw
}

// buildFailureLinks builds failure links using BFS.
func (m *ACMatcher) buildFailureLinks() error {
	queue := []*acNode{}

	// Set fail links for depth-1 nodes to root
	for _, child := range m.root.children {
		child.fail = m.root
		queue = append(queue, child)
	}

	// BFS
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for r, child := range current.children {
			queue = append(queue, child)

			// Find fail link for this child
			fail := current.fail
			for fail != m.root && fail.children[r] == nil {
				fail = fail.fail
			}

			if fail.children[r] != nil && fail.children[r] != child {
				child.fail = fail.children[r]
			} else {
				child.fail = m.root
			}

			// Propagate output flag
			if child.fail.output {
				child.output = true
			}
		}
	}

	return nil
}
