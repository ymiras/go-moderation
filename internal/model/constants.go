package model

// ActionType represents the moderation action to take
type ActionType string

const (
	ActionPass  ActionType = "pass"  // Allow content
	ActionBlock ActionType = "block" // Block content
	ActionMask  ActionType = "mask"  // Mask sensitive content
)

// PointType represents the moderation point (input or output)
type PointType string

const (
	PointInput  PointType = "input"  // Input content moderation
	PointOutput PointType = "output" // Output content moderation
)

// Severity represents the severity level of a hit
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)
