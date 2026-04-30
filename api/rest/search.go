package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/funcman/back-pushing/internal/engine/search"
	"github.com/funcman/back-pushing/internal/storage/memory"
)

type SearchHandler struct {
	filterEngine  *search.FilterEngine
	fulltextIndex *search.FullTextIndex
}

func NewSearchHandler(store *memory.ObjectStore) *SearchHandler {
	return &SearchHandler{
		filterEngine:  search.NewFilterEngine(store),
		fulltextIndex:  search.NewFullTextIndex(),
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
		Type  string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Query != "" {
		results, _ := h.fulltextIndex.Search(c.Request.Context(), req.Query)
		c.JSON(http.StatusOK, gin.H{"query": req.Query, "results": results})
		return
	}

	c.JSON(http.StatusOK, gin.H{"query": req.Query, "results": []any{}})
}