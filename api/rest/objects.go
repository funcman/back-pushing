package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/funcman/back-pushing/internal/storage/memory"
)

type ObjectsHandler struct {
	store *memory.ObjectStore
}

func NewObjectsHandler(store *memory.ObjectStore) *ObjectsHandler {
	return &ObjectsHandler{store: store}
}

func (h *ObjectsHandler) Get(c *gin.Context) {
	objType := c.Param("type")
	id := c.Param("id")

	obj, err := h.store.Get(c.Request.Context(), objType, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, obj)
}

func (h *ObjectsHandler) Create(c *gin.Context) {
	var req struct {
		Type string         `json:"type"`
		ID   string         `json:"id"`
		Data map[string]any `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.Create(c.Request.Context(), req.Type, req.ID, req.Data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": req.ID})
}

func (h *ObjectsHandler) Update(c *gin.Context) {
	objType := c.Param("type")
	id := c.Param("id")

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.Update(c.Request.Context(), objType, id, data); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ObjectsHandler) Delete(c *gin.Context) {
	objType := c.Param("type")
	id := c.Param("id")

	if err := h.store.Delete(c.Request.Context(), objType, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *ObjectsHandler) List(c *gin.Context) {
	objType := c.Param("type")

	objects, err := h.store.List(c.Request.Context(), objType, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, objects)
}