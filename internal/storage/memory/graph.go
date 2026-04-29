package memory

import (
	"context"
	"sync"

	"github.com/back-pushing/back-pushing/internal/storage"
)

type GraphStore struct {
	mu     sync.RWMutex
	edges  map[string]map[string][]storage.Edge
	revIdx map[string]map[string][]storage.Edge
}

func NewGraphStore() *GraphStore {
	return &GraphStore{
		edges:  make(map[string]map[string][]storage.Edge),
		revIdx: make(map[string]map[string][]storage.Edge),
	}
}

func (s *GraphStore) removeEdgeFromSlice(edges []storage.Edge, from, to string) []storage.Edge {
	for i, e := range edges {
		if e.From == from && e.To == to {
			return append(edges[:i], edges[i+1:]...)
		}
	}
	return edges
}

func (s *GraphStore) AddEdge(ctx context.Context, linkType string, fromID, toID string, props map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.edges[linkType] == nil {
		s.edges[linkType] = make(map[string][]storage.Edge)
	}
	if s.revIdx[linkType] == nil {
		s.revIdx[linkType] = make(map[string][]storage.Edge)
	}

	edge := storage.Edge{
		From:  fromID,
		To:    toID,
		Props: props,
	}

	s.edges[linkType][fromID] = append(s.edges[linkType][fromID], edge)
	s.revIdx[linkType][toID] = append(s.revIdx[linkType][toID], edge)

	return nil
}

func (s *GraphStore) RemoveEdge(ctx context.Context, linkType string, fromID, toID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.edges[linkType] == nil {
		return storage.ErrNotFound
	}

	s.edges[linkType][fromID] = s.removeEdgeFromSlice(s.edges[linkType][fromID], fromID, toID)
	s.revIdx[linkType][toID] = s.removeEdgeFromSlice(s.revIdx[linkType][toID], fromID, toID)

	return nil
}

func (s *GraphStore) GetEdges(ctx context.Context, linkType string, nodeID string) ([]storage.Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.edges[linkType] == nil {
		return []storage.Edge{}, nil
	}

	var result []storage.Edge

	if edges, ok := s.edges[linkType][nodeID]; ok {
		result = append(result, edges...)
	}

	if revEdges, ok := s.revIdx[linkType][nodeID]; ok {
		result = append(result, revEdges...)
	}

	return result, nil
}

func (s *GraphStore) GetOutgoingEdges(ctx context.Context, linkType string, nodeID string) ([]storage.Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.edges[linkType][nodeID], nil
}

func (s *GraphStore) GetIncomingEdges(ctx context.Context, linkType string, nodeID string) ([]storage.Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revIdx[linkType][nodeID], nil
}

func (s *GraphStore) Traverse(ctx context.Context, startID string, linkTypes []string, depth int) ([]storage.Path, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var paths []storage.Path
	visited := make(map[string]bool)

	s.dfs(startID, linkTypes, depth, visited, nil, nil, &paths)

	return paths, nil
}

func (s *GraphStore) dfs(currentID string, linkTypes []string, depth int, visited map[string]bool, nodes []string, edges []storage.Edge, paths *[]storage.Path) {
	if depth < 0 || visited[currentID] {
		return
	}

	visited[currentID] = true
	nodes = append(nodes, currentID)

	if len(edges) > 0 {
		path := storage.Path{
			Nodes: make([]string, len(nodes)),
			Edges: make([]storage.Edge, len(edges)),
		}
		copy(path.Nodes, nodes)
		copy(path.Edges, edges)
		*paths = append(*paths, path)
	}

	for _, linkType := range linkTypes {
		if s.edges[linkType] == nil {
			continue
		}
		for _, edge := range s.edges[linkType][currentID] {
			newEdges := make([]storage.Edge, len(edges)+1)
			copy(newEdges, edges)
			newEdges[len(edges)] = edge

			newVisited := make(map[string]bool)
			for k, v := range visited {
				newVisited[k] = v
			}

			s.dfs(edge.To, linkTypes, depth-1, newVisited, nodes, newEdges, paths)
		}
	}
}
