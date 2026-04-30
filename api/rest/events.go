package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/funcman/back-pushing/internal/engine/temporal"
)

type EventsHandler struct {
	analyzer *temporal.TemporalAnalyzer
}

func NewEventsHandler(analyzer *temporal.TemporalAnalyzer) *EventsHandler {
	return &EventsHandler{analyzer: analyzer}
}

func (h *EventsHandler) Record(c *gin.Context) {
	var req struct {
		Type  string         `json:"type"`
		Actor string         `json:"actor"`
		Props map[string]any `json:"props"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := temporal.Event{
		Type:      req.Type,
		Actor:     req.Actor,
		Props:     req.Props,
		Timestamp: time.Now(),
	}

	h.analyzer.RecordEvent(c.Request.Context(), event)
	c.JSON(http.StatusCreated, gin.H{"status": "recorded"})
}

func (h *EventsHandler) GetByActor(c *gin.Context) {
	objType := c.Param("type")
	actor := c.Param("id")

	events, _ := h.analyzer.GetEvents(c.Request.Context(), objType, actor, time.Time{}, time.Now())
	c.JSON(http.StatusOK, events)
}