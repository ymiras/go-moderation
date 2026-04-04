package dify

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/dify-moderation/internal/engine"
	"github.com/ymiras/dify-moderation/internal/model"
)

// Handler handles Dify Moderation API Extension requests.
type Handler struct {
	svc *engine.ModerationService
}

// NewHandler creates a new Dify handler.
func NewHandler(svc *engine.ModerationService) *Handler {
	return &Handler{svc: svc}
}

// Request represents a Dify Moderation API Extension request.
type Request struct {
	Point  string `json:"point" binding:"required"`
	Params Params `json:"params" binding:"required"`
}

// Params represents the parameters in a Dify request.
type Params struct {
	AppID  string         `json:"app_id"`
	Inputs map[string]any `json:"inputs"`
	Query  string         `json:"query"`
	Text   string         `json:"text"`
}

// Response represents a Dify Moderation API Extension response.
type Response struct {
	Flagged        bool   `json:"flagged"`
	Action         string `json:"action"`
	PresetResponse string `json:"preset_response,omitempty"`
}

// Moderate handles POST /dify/moderation.
func (h *Handler) Moderate(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate point type
	if req.Point != "app.moderation.input" && req.Point != "app.moderation.output" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported point type: " + req.Point})
		return
	}

	// Extract text to moderate
	text := ""
	point := model.PointInput
	if req.Point == "app.moderation.input" {
		text = req.Params.Query
		point = model.PointInput
	} else {
		text = req.Params.Text
		point = model.PointOutput
	}

	// Validate text is provided
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	// Create moderation context
	ctx := &engine.ModerationContext{
		Text:  text,
		Point: point,
		AppID: req.Params.AppID,
	}

	// Call moderation service
	result, err := h.svc.Moderate(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "moderation failed"})
		return
	}

	// Build response
	if result.Flagged {
		c.JSON(http.StatusOK, Response{
			Flagged:        true,
			Action:         "direct_output",
			PresetResponse: "内容因违反政策被拦截",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Flagged: false,
		Action:  "direct_output",
	})
}
