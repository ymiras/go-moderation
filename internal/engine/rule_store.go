package engine

// RuleStore is the interface for rule storage.
type RuleStore interface {
	GetRules() []Rule
}
