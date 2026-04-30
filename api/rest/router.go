package rest

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine  *gin.Engine
	objects *ObjectsHandler
	search  *SearchHandler
	events  *EventsHandler
}

func NewRouter(objects *ObjectsHandler, search *SearchHandler, events *EventsHandler) *Router {
	r := &Router{
		engine:  gin.Default(),
		objects: objects,
		search:  search,
		events:  events,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	api := r.engine.Group("/api")
	{
		// Objects
		api.GET("/objects/:type/:id", r.objects.Get)
		api.POST("/objects", r.objects.Create)
		api.PUT("/objects/:type/:id", r.objects.Update)
		api.DELETE("/objects/:type/:id", r.objects.Delete)
		api.GET("/objects/:type", r.objects.List)

		// Search
		api.POST("/search", r.search.Search)

		// Events
		api.POST("/events", r.events.Record)
		api.GET("/events/:type/:id", r.events.GetByActor)
	}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}