package search

import (
	"context"

	"github.com/funcman/back-pushing/internal/storage/memory"
)

type AggregationResult struct {
	Count  int64
	Sum    map[string]float64
	Avg    map[string]float64
	Min    map[string]float64
	Max    map[string]float64
	Groups map[string][]map[string]any
}

type AggregateEngine struct {
	store *memory.ObjectStore
}

func NewAggregateEngine(store *memory.ObjectStore) *AggregateEngine {
	return &AggregateEngine{store: store}
}

func (e *AggregateEngine) Aggregate(ctx context.Context, objType string, fields []string, groupBy string) (*AggregationResult, error) {
	objects, err := e.store.List(ctx, objType, nil)
	if err != nil {
		return nil, err
	}

	result := &AggregationResult{
		Sum:    make(map[string]float64),
		Avg:    make(map[string]float64),
		Min:    make(map[string]float64),
		Max:    make(map[string]float64),
		Groups: make(map[string][]map[string]any),
	}

	for _, obj := range objects {
		result.Count++

		groupKey := ""
		if groupBy != "" {
			if v, ok := obj[groupBy].(string); ok {
				groupKey = v
			}
		}

		for _, field := range fields {
			if v, ok := obj[field].(float64); ok {
				result.Sum[field] += v
				if result.Min[field] == 0 || v < result.Min[field] {
					result.Min[field] = v
				}
				if v > result.Max[field] {
					result.Max[field] = v
				}
			}
		}

		if groupKey != "" {
			result.Groups[groupKey] = append(result.Groups[groupKey], obj)
		}
	}

	for _, field := range fields {
		if result.Count > 0 {
			result.Avg[field] = result.Sum[field] / float64(result.Count)
		}
	}

	return result, nil
}