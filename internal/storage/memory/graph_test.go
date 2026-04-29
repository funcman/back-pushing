package memory

import (
	"context"
	"testing"
)

func TestGraphStore_AddEdge(t *testing.T) {
	store := NewGraphStore()
	ctx := context.Background()

	err := store.AddEdge(ctx, "knows", "user1", "user2", map[string]any{"since": "2020"})
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	edges, err := store.GetEdges(ctx, "knows", "user1")
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}

	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0].From != "user1" || edges[0].To != "user2" {
		t.Errorf("unexpected edge: %v", edges[0])
	}
}

func TestGraphStore_RemoveEdge(t *testing.T) {
	store := NewGraphStore()
	ctx := context.Background()

	store.AddEdge(ctx, "knows", "user1", "user2", nil)

	err := store.RemoveEdge(ctx, "knows", "user1", "user2")
	if err != nil {
		t.Fatalf("RemoveEdge failed: %v", err)
	}

	edges, err := store.GetEdges(ctx, "knows", "user1")
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}

	found := false
	for _, e := range edges {
		if e.From == "user1" && e.To == "user2" {
			found = true
			break
		}
	}
	if found {
		t.Error("edge should have been removed")
	}
}

func TestGraphStore_GetEdges(t *testing.T) {
	store := NewGraphStore()
	ctx := context.Background()

	store.AddEdge(ctx, "knows", "user1", "user2", nil)
	store.AddEdge(ctx, "knows", "user2", "user3", nil)
	store.AddEdge(ctx, "knows", "user1", "user3", nil)
	store.AddEdge(ctx, "follows", "user1", "user2", nil)

	t.Run("Get outgoing edges", func(t *testing.T) {
		edges, err := store.GetEdges(ctx, "knows", "user1")
		if err != nil {
			t.Fatalf("GetEdges failed: %v", err)
		}
		if len(edges) != 2 {
			t.Errorf("expected 2 edges from user1, got %d", len(edges))
		}
	})

	t.Run("Get incoming edges", func(t *testing.T) {
		edges, err := store.GetEdges(ctx, "knows", "user3")
		if err != nil {
			t.Fatalf("GetEdges failed: %v", err)
		}
		if len(edges) != 2 {
			t.Errorf("expected 2 edges to user3, got %d", len(edges))
		}
	})

	t.Run("Get non-existent link type", func(t *testing.T) {
		edges, err := store.GetEdges(ctx, "nonexistent", "user1")
		if err != nil {
			t.Fatalf("GetEdges failed: %v", err)
		}
		if len(edges) != 0 {
			t.Errorf("expected 0 edges, got %d", len(edges))
		}
	})
}

func TestGraphStore_Traverse(t *testing.T) {
	store := NewGraphStore()
	ctx := context.Background()

	store.AddEdge(ctx, "knows", "user1", "user2", nil)
	store.AddEdge(ctx, "knows", "user2", "user3", nil)
	store.AddEdge(ctx, "knows", "user3", "user4", nil)
	store.AddEdge(ctx, "knows", "user1", "user3", nil)

	t.Run("Traverse depth 1", func(t *testing.T) {
		paths, err := store.Traverse(ctx, "user1", []string{"knows"}, 1)
		if err != nil {
			t.Fatalf("Traverse failed: %v", err)
		}
		if len(paths) != 2 {
			t.Errorf("expected 2 paths, got %d", len(paths))
		}
	})

	t.Run("Traverse depth 2", func(t *testing.T) {
		paths, err := store.Traverse(ctx, "user1", []string{"knows"}, 2)
		if err != nil {
			t.Fatalf("Traverse failed: %v", err)
		}
		if len(paths) == 0 {
			t.Error("expected paths at depth 2")
		}
	})

	t.Run("Traverse non-existent start", func(t *testing.T) {
		paths, err := store.Traverse(ctx, "nonexistent", []string{"knows"}, 2)
		if err != nil {
			t.Fatalf("Traverse failed: %v", err)
		}
		if len(paths) != 0 {
			t.Errorf("expected 0 paths, got %d", len(paths))
		}
	})
}