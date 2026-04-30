package search

import (
	"context"
	"testing"

	"github.com/funcman/back-pushing/internal/storage/memory"
)

func TestFullTextIndex_Search(t *testing.T) {
	idx := NewFullTextIndex()
	ctx := context.Background()

	idx.Index(ctx, "Person", "p1", map[string]any{"name": "Alice Smith"})
	idx.Index(ctx, "Person", "p2", map[string]any{"name": "Bob Johnson"})
	idx.Index(ctx, "Person", "p3", map[string]any{"name": "Alice Williams"})

	results, err := idx.Search(ctx, "Alice")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFilterEngine_Filter(t *testing.T) {
	store := memory.NewObjectStore()
	engine := NewFilterEngine(store)
	ctx := context.Background()

	store.Create(ctx, "Person", "p1", map[string]any{"name": "Alice", "age": float64(30)})
	store.Create(ctx, "Person", "p2", map[string]any{"name": "Bob", "age": float64(25)})
	store.Create(ctx, "Person", "p3", map[string]any{"name": "Charlie", "age": float64(35)})

	filter := NewFilter().GreaterThan("age", float64(28))
	results, err := engine.Filter(ctx, "Person", filter)
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}