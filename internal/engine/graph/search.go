package graph

import (
	"context"
	"strings"

	"github.com/funcman/back-pushing/internal/storage/memory"
)

type SearchResult struct {
	NodeID string
	Type   string
	Data   map[string]any
}

type SearchEngine struct {
	traversal *TraversalEngine
	objStore  *memory.ObjectStore
}

func NewSearchEngine(t *TraversalEngine, o *memory.ObjectStore) *SearchEngine {
	return &SearchEngine{
		traversal: t,
		objStore:  o,
	}
}

func (e *SearchEngine) FindConnected(ctx context.Context, nodeID string, linkName string, maxDepth int) ([]SearchResult, error) {
	connected, err := e.traversal.BFS(ctx, nodeID, linkName, maxDepth)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, id := range connected {
		if id == nodeID {
			continue
		}
		results = append(results, SearchResult{
			NodeID: id,
			Type:   linkName,
		})
	}

	return results, nil
}

func (e *SearchEngine) FullTextSearch(ctx context.Context, objType string, query string) ([]SearchResult, error) {
	all, err := e.objStore.List(ctx, objType, nil)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, obj := range all {
		for _, v := range obj {
			if str, ok := v.(string); ok {
				if strings.Contains(strings.ToLower(str), query) {
					results = append(results, SearchResult{
						NodeID: getID(obj),
						Type:   objType,
						Data:   obj,
					})
					break
				}
			}
		}
	}

	return results, nil
}

func getID(data map[string]any) string {
	if id, ok := data["id"].(string); ok {
		return id
	}
	return ""
}
