package model

// Keyword represents a keyword in the word bank
type Keyword struct {
	Word     string
	Type     string
	Severity Severity
	Action   ActionType
}

// HitRecord represents a single match found during moderation
type HitRecord struct {
	Word     string
	Type     string
	Severity Severity
	Index    int // position in text
	Length   int // match length
}

// ModerationResult represents the result of a moderation check
type ModerationResult struct {
	Flagged   bool
	Action    ActionType
	Hits      []HitRecord
	LatencyMs float64
}
