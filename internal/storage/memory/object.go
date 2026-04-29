package memory

import (
	"context"
	"sync"

	"github.com/back-pushing/back-pushing/internal/storage"
)

type ObjectStore struct {
	mu    sync.RWMutex
	store map[string]map[string]map[string]any
}

func NewObjectStore() *ObjectStore {
	return &ObjectStore{
		store: make(map[string]map[string]map[string]any),
	}
}

func (s *ObjectStore) ensureType(objType string) {
	if s.store[objType] == nil {
		s.store[objType] = make(map[string]map[string]any)
	}
}

func (s *ObjectStore) matchFilter(data map[string]any, filter map[string]any) bool {
	for k, v := range filter {
		if data[k] != v {
			return false
		}
	}
	return true
}

func (s *ObjectStore) Create(ctx context.Context, objType string, id string, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureType(objType)

	if s.store[objType][id] != nil {
		return storage.ErrAlreadyExists
	}

	s.store[objType][id] = data
	return nil
}

func (s *ObjectStore) Get(ctx context.Context, objType string, id string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.store[objType] == nil {
		return nil, storage.ErrNotFound
	}

	obj, ok := s.store[objType][id]
	if !ok {
		return nil, storage.ErrNotFound
	}

	result := make(map[string]any)
	for k, v := range obj {
		result[k] = v
	}
	return result, nil
}

func (s *ObjectStore) Update(ctx context.Context, objType string, id string, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store[objType] == nil || s.store[objType][id] == nil {
		return storage.ErrNotFound
	}

	s.store[objType][id] = data
	return nil
}

func (s *ObjectStore) Delete(ctx context.Context, objType string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store[objType] == nil {
		return storage.ErrNotFound
	}

	if s.store[objType][id] == nil {
		return storage.ErrNotFound
	}

	delete(s.store[objType], id)
	return nil
}

func (s *ObjectStore) List(ctx context.Context, objType string, filter map[string]any) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.store[objType] == nil {
		return []map[string]any{}, nil
	}

	var result []map[string]any
	for _, data := range s.store[objType] {
		if filter == nil || s.matchFilter(data, filter) {
			obj := make(map[string]any)
			for k, v := range data {
				obj[k] = v
			}
			result = append(result, obj)
		}
	}

	return result, nil
}
