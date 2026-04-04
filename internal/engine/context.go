package engine

import "github.com/ymiras/go-moderation/internal/model"

// ModerationContext carries the input data for a moderation request.
// It is a pure data struct with no dependencies.
type ModerationContext struct {
	// Text is the content to be moderated.
	Text string

	// Point indicates whether this is input or output moderation.
	Point model.PointType

	// AppID is an optional application identifier.
	AppID string
}
