package memory

import (
	"context"
	"testing"

	"github.com/funcman/back-pushing/internal/storage"
)

func TestObjectStore_CRUD(t *testing.T) {
	store := NewObjectStore()
	ctx := context.Background()

	t.Run("Create and Get", func(t *testing.T) {
		data := map[string]any{"name": "test", "value": 42}
		err := store.Create(ctx, "user", "user1", data)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		result, err := store.Get(ctx, "user", "user1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name=test, got %v", result["name"])
		}
		if result["value"] != 42 {
			t.Errorf("expected value=42, got %v", result["value"])
		}
	})

	t.Run("Create duplicate returns error", func(t *testing.T) {
		data := map[string]any{"name": "test2"}
		err := store.Create(ctx, "user", "user1", data)
		if err != storage.ErrAlreadyExists {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("Get non-existent returns error", func(t *testing.T) {
		_, err := store.Get(ctx, "user", "nonexistent")
		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		newData := map[string]any{"name": "updated", "score": 100}
		err := store.Update(ctx, "user", "user1", newData)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		result, err := store.Get(ctx, "user", "user1")
		if err != nil {
			t.Fatalf("Get after Update failed: %v", err)
		}

		if result["name"] != "updated" {
			t.Errorf("expected name=updated, got %v", result["name"])
		}
		if result["score"] != 100 {
			t.Errorf("expected score=100, got %v", result["score"])
		}
	})

	t.Run("Update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, "user", "nonexistent", map[string]any{"name": "test"})
		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(ctx, "user", "user1")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = store.Get(ctx, "user", "user1")
		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("Delete non-existent returns error", func(t *testing.T) {
		err := store.Delete(ctx, "user", "nonexistent")
		if err != storage.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestObjectStore_List(t *testing.T) {
	store := NewObjectStore()
	ctx := context.Background()

	store.Create(ctx, "user", "user1", map[string]any{"name": "alice", "role": "admin"})
	store.Create(ctx, "user", "user2", map[string]any{"name": "bob", "role": "user"})
	store.Create(ctx, "user", "user3", map[string]any{"name": "charlie", "role": "admin"})
	store.Create(ctx, "product", "prod1", map[string]any{"name": "widget"})

	t.Run("List all users", func(t *testing.T) {
		results, err := store.List(ctx, "user", nil)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 users, got %d", len(results))
		}
	})

	t.Run("List with filter", func(t *testing.T) {
		results, err := store.List(ctx, "user", map[string]any{"role": "admin"})
		if err != nil {
			t.Fatalf("List with filter failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 admins, got %d", len(results))
		}
	})

	t.Run("List non-existent type", func(t *testing.T) {
		results, err := store.List(ctx, "nonexistent", nil)
		if err != nil {
			t.Fatalf("List nonexistent type failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}
