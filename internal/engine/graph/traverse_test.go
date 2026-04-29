package graph

import (
	"context"
	"testing"

	"github.com/back-pushing/back-pushing/internal/storage/memory"
)

func TestTraversalEngine_BFS(t *testing.T) {
	graph := memory.NewGraphStore()
	ctx := context.Background()

	// Build test graph:
	//   A -> B -> C
	//   A -> D -> E
	//   B -> F

	graph.AddEdge(ctx, "knows", "A", "B", nil)
	graph.AddEdge(ctx, "knows", "B", "C", nil)
	graph.AddEdge(ctx, "knows", "A", "D", nil)
	graph.AddEdge(ctx, "knows", "D", "E", nil)
	graph.AddEdge(ctx, "knows", "B", "F", nil)

	engine := NewTraversalEngine(graph)

	// Test BFS from A with depth 2
	result, err := engine.BFS(ctx, "A", "knows", 2)
	if err != nil {
		t.Fatalf("BFS failed: %v", err)
	}

	// A should be included, and at depth 1: B, D; at depth 2: C, E, F
	// Order may vary due to map iteration, so check membership
	expected := map[string]bool{"A": true, "B": true, "D": true, "C": true, "E": true, "F": true}
	for _, id := range result {
		if !expected[id] {
			t.Errorf("unexpected id in result: %s", id)
		}
		delete(expected, id)
	}
	for id := range expected {
		t.Errorf("missing expected id: %s", id)
	}
}

func TestSearchEngine_FullTextSearch(t *testing.T) {
	graph := memory.NewGraphStore()
	objStore := memory.NewObjectStore()
	ctx := context.Background()

	// Create test objects
	objStore.Create(ctx, "person", "p1", map[string]any{"id": "p1", "name": "Alice", "bio": "Likes hiking"})
	objStore.Create(ctx, "person", "p2", map[string]any{"id": "p2", "name": "Bob", "bio": "Likes coding"})
	objStore.Create(ctx, "person", "p3", map[string]any{"id": "p3", "name": "Charlie", "bio": "Likes hiking and coding"})

	traversal := NewTraversalEngine(graph)
	search := NewSearchEngine(traversal, objStore)

	// Search for "hiking"
	results, err := search.FullTextSearch(ctx, "person", "hiking")
	if err != nil {
		t.Fatalf("FullTextSearch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	expectedIDs := map[string]bool{"p1": true, "p3": true}
	for _, r := range results {
		if !expectedIDs[r.NodeID] {
			t.Errorf("unexpected result id: %s", r.NodeID)
		}
		if r.Type != "person" {
			t.Errorf("expected type 'person', got %s", r.Type)
		}
	}

	// Search for "coding"
	results, err = search.FullTextSearch(ctx, "person", "coding")
	if err != nil {
		t.Fatalf("FullTextSearch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for 'coding', got %d", len(results))
	}
}