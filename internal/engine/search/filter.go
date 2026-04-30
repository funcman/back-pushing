package search

import (
	"context"

	"github.com/funcman/back-pushing/internal/storage/memory"
)

type Filter struct {
	Eq  map[string]any
	Gt  map[string]any
	Lt  map[string]any
	Gte map[string]any
	Lte map[string]any
	In  map[string][]any
}

func NewFilter() *Filter {
	return &Filter{
		Eq:  make(map[string]any),
		Gt:  make(map[string]any),
		Lt:  make(map[string]any),
		Gte: make(map[string]any),
		Lte: make(map[string]any),
		In:  make(map[string][]any),
	}
}

func (f *Filter) Equal(key string, value any) *Filter {
	f.Eq[key] = value
	return f
}

func (f *Filter) GreaterThan(key string, value any) *Filter {
	f.Gt[key] = value
	return f
}

func (f *Filter) LessThan(key string, value any) *Filter {
	f.Lt[key] = value
	return f
}

func (f *Filter) InList(key string, values []any) *Filter {
	f.In[key] = values
	return f
}

type FilterEngine struct {
	store *memory.ObjectStore
}

func NewFilterEngine(store *memory.ObjectStore) *FilterEngine {
	return &FilterEngine{store: store}
}

func (e *FilterEngine) Filter(ctx context.Context, objType string, filter *Filter) ([]map[string]any, error) {
	objects, err := e.store.List(ctx, objType, nil)
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for _, obj := range objects {
		if e.matches(obj, filter) {
			result = append(result, obj)
		}
	}
	return result, nil
}

func (e *FilterEngine) matches(obj map[string]any, filter *Filter) bool {
	for key, value := range filter.Eq {
		if obj[key] != value {
			return false
		}
	}

	for key, value := range filter.Gt {
		if !compareGT(obj[key], value) {
			return false
		}
	}

	for key, value := range filter.Lt {
		if !compareLT(obj[key], value) {
			return false
		}
	}

	for key, values := range filter.In {
		if !contains(obj[key], values) {
			return false
		}
	}

	return true
}

func compareGT(a, b any) bool {
	af, ok := a.(float64)
	bf, ok2 := b.(float64)
	if ok && ok2 {
		return af > bf
	}
	return false
}

func compareLT(a, b any) bool {
	af, ok := a.(float64)
	bf, ok2 := b.(float64)
	if ok && ok2 {
		return af < bf
	}
	return false
}

func contains(a any, list []any) bool {
	for _, v := range list {
		if a == v {
			return true
		}
	}
	return false
}