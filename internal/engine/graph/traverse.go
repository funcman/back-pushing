package graph

import (
	"context"

	"github.com/back-pushing/back-pushing/internal/storage/memory"
	"github.com/back-pushing/back-pushing/internal/storage"
)

// Path is an alias for storage.Path
type Path = storage.Path

type TraversalEngine struct {
	graph   *memory.GraphStore
	linkMap map[string]string // link name -> actual link type
}

func NewTraversalEngine(g *memory.GraphStore) *TraversalEngine {
	return &TraversalEngine{
		graph:   g,
		linkMap: make(map[string]string),
	}
}

func (e *TraversalEngine) RegisterLink(name, linkType string) {
	e.linkMap[name] = linkType
}

func (e *TraversalEngine) BFS(ctx context.Context, startID string, linkName string, maxDepth int) ([]string, error) {
	linkType := e.linkMap[linkName]
	if linkType == "" {
		linkType = linkName
	}

	var result []string
	visited := make(map[string]bool)
	queue := []string{startID}
	currentDepth := 0

	for len(queue) > 0 && currentDepth <= maxDepth {
		var nextQueue []string
		for _, id := range queue {
			if visited[id] {
				continue
			}
			visited[id] = true
			result = append(result, id)

			edges, err := e.graph.GetOutgoingEdges(ctx, linkType, id)
			if err != nil {
				return nil, err
			}
			for _, edge := range edges {
				if !visited[edge.To] {
					nextQueue = append(nextQueue, edge.To)
				}
			}
		}
		queue = nextQueue
		currentDepth++
	}

	return result, nil
}

func (e *TraversalEngine) FindPaths(ctx context.Context, startID string, linkNames []string, maxDepth int) ([]Path, error) {
	linkTypes := make([]string, len(linkNames))
	for i, name := range linkNames {
		if lt, ok := e.linkMap[name]; ok {
			linkTypes[i] = lt
		} else {
			linkTypes[i] = name
		}
	}

	return e.graph.Traverse(ctx, startID, linkTypes, maxDepth)
}