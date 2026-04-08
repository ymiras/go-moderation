package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/engine"
	"github.com/ymiras/go-moderation/internal/model"
)

// Handler handles Standard REST API moderation requests.
type Handler struct {
	svc *engine.ModerationService
}

// NewHandler creates a new Standard handler.
func NewHandler(svc *engine.ModerationService) *Handler {
	return &Handler{svc: svc}
}

// Request represents a Standard REST API moderation request.
type Request struct {
	Text  string `json:"text" binding:"required"`
	Point string `json:"point"`
	AppID string `json:"app_id"`
}

// Response represents a Standard REST API moderation response.
type Response struct {
	Flagged   bool              `json:"flagged"`
	Action    model.ActionType  `json:"action"`
	Hits      []model.HitRecord `json:"hits"`
	LatencyMs float64           `json:"latency_ms"`
}

// Moderate handles POST /api/moderate.
func (h *Handler) Moderate(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	// Validate point
	point := model.PointInput
	if req.Point == "output" {
		point = model.PointOutput
	} else if req.Point != "" && req.Point != "input" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "point must be 'input' or 'output'"})
		return
	}

	// Create moderation context
	ctx := &engine.ModerationContext{
		Text:  req.Text,
		Point: point,
		AppID: req.AppID,
	}

	// Call moderation service
	result, err := h.svc.Moderate(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "moderation failed"})
		return
	}

	// Build response
	c.JSON(http.StatusOK, Response{
		Flagged:   result.Flagged,
		Action:    result.Action,
		Hits:      result.Hits,
		LatencyMs: result.LatencyMs,
	})
}
