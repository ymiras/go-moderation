package matcher

import (
	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// Matcher defines the interface for content matching algorithms.
type Matcher interface {
	// Name returns the unique name of this matcher for logging and debugging.
	Name() string

	// Match searches for keywords in the given text using the word bank.
	// Returns all hit records found, or an error if the matching fails.
	Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error)
}
