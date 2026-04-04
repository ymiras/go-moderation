package engine

import (
	"testing"

	"github.com/ymiras/dify-moderation/internal/model"
)

func TestModerationContext(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		ctx := &ModerationContext{
			Text:  "hello world",
			Point: model.PointInput,
			AppID: "app-123",
		}

		if ctx.Text != "hello world" {
			t.Errorf("expected Text='hello world', got '%s'", ctx.Text)
		}
		if ctx.Point != model.PointInput {
			t.Errorf("expected Point=PointInput, got %v", ctx.Point)
		}
		if ctx.AppID != "app-123" {
			t.Errorf("expected AppID='app-123', got '%s'", ctx.AppID)
		}
	})

	t.Run("empty fields allowed", func(t *testing.T) {
		ctx := &ModerationContext{}

		if ctx.Text != "" {
			t.Errorf("expected empty Text, got '%s'", ctx.Text)
		}
		if ctx.Point != "" {
			t.Errorf("expected empty Point, got %v", ctx.Point)
		}
		if ctx.AppID != "" {
			t.Errorf("expected empty AppID, got '%s'", ctx.AppID)
		}
	})
}
