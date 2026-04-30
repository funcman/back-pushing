package temporal

import (
	"context"
	"sort"
	"time"
)

type Event struct {
	ID        string
	Type      string
	Timestamp time.Time
	Actor     string
	Props     map[string]any
}

type TemporalAnalyzer struct {
	events []Event
}

func NewTemporalAnalyzer() *TemporalAnalyzer {
	return &TemporalAnalyzer{
		events: []Event{},
	}
}

func (a *TemporalAnalyzer) RecordEvent(ctx context.Context, event Event) error {
	a.events = append(a.events, event)
	sort.Slice(a.events, func(i, j int) bool {
		return a.events[i].Timestamp.Before(a.events[j].Timestamp)
	})
	return nil
}

func (a *TemporalAnalyzer) GetEvents(ctx context.Context, objType, objID string, since, until time.Time) ([]Event, error) {
	var result []Event
	for _, e := range a.events {
		if e.Type == objType && e.Actor == objID {
			if since.IsZero() || e.Timestamp.After(since) {
				if until.IsZero() || e.Timestamp.Before(until) {
					result = append(result, e)
				}
			}
		}
	}
	return result, nil
}

func (a *TemporalAnalyzer) GetEventCount(ctx context.Context, objType, objID string, since, until time.Time) (int64, error) {
	events, err := a.GetEvents(ctx, objType, objID, since, until)
	if err != nil {
		return 0, err
	}
	return int64(len(events)), nil
}
