package engine

import "github.com/ymiras/go-moderation/internal/model"

// Decision represents the result of a rule engine evaluation.
type Decision struct {
	Action    model.ActionType
	Hits      []model.HitRecord
	RuleID    string
	LatencyMs float64
}
