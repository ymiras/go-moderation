package engine

import "github.com/ymiras/go-moderation/internal/model"

// Rule defines a single moderation rule.
type Rule struct {
	// ID is the unique identifier for the rule.
	ID string
	// Name is a human-readable name for the rule.
	Name string
	// Matcher is the name of the matcher to use.
	Matcher string
	// Condition is the condition to evaluate (暂用 interface{}，后续 condition-system 替换)
	Condition interface{}
	// Action is the action to take when the rule matches.
	Action model.ActionType
	// Priority determines the order of evaluation (higher = more important).
	Priority int
}
