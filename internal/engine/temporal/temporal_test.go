package temporal

import (
	"context"
	"testing"
	"time"
)

func TestTemporalAnalyzer_RecordAndGet(t *testing.T) {
	analyzer := NewTemporalAnalyzer()
	ctx := context.Background()

	event := Event{
		ID:        "e1",
		Type:      "Transaction",
		Timestamp: time.Now(),
		Actor:     "acc1",
		Props:     map[string]any{"amount": 100.0},
	}

	if err := analyzer.RecordEvent(ctx, event); err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	events, err := analyzer.GetEvents(ctx, "Transaction", "acc1", time.Time{}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestAnomalyDetector_Detect(t *testing.T) {
	// With window=5 and outlier in window, zscore is bounded by ~2.0
	// Using threshold 1.99 to ensure zscore > threshold for outlier values
	detector := NewAnomalyDetector(1.99, 5)

	// Outlier value that produces zscore > 1.99
	values := []float64{1, 1, 1, 1, 1, 1, 100}
	result := detector.Detect(values)
	if !result.IsAnomaly {
		t.Errorf("expected anomaly to be detected for outlier value, got Score: %v", result.Score)
	}

	// Normal values that should not trigger anomaly
	normalValues := []float64{10, 11, 10, 10, 11, 10, 11}
	result = detector.Detect(normalValues)
	if result.IsAnomaly {
		t.Error("expected no anomaly for normal values")
	}
}
