package engine

import "sync"

// MemoryRuleStore is an in-memory implementation of RuleStore.
type MemoryRuleStore struct {
	mu    sync.RWMutex
	rules []Rule
}

// NewMemoryRuleStore creates a new MemoryRuleStore.
func NewMemoryRuleStore() *MemoryRuleStore {
	return &MemoryRuleStore{
		rules: []Rule{},
	}
}

// AddRules adds rules to the store.
func (s *MemoryRuleStore) AddRules(rules []Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, rules...)
}

// GetRules returns all rules.
func (s *MemoryRuleStore) GetRules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules
}
