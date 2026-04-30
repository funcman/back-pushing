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
	// SPEC: threshold=2.0, window=5
	detector := NewAnomalyDetector(2.0, 5)

	// Outlier test: baseline={10,11,10,10,11}, latest=100
	// Mean=10.4, stddev≈0.49, z-score≈182.9 >> 2.0 → anomaly detected
	values := []float64{10, 11, 10, 10, 11, 10, 100.0}
	result := detector.Detect(values)
	if !result.IsAnomaly {
		t.Errorf("expected anomaly to be detected for outlier value, got Score: %v", result.Score)
	}

	// Normal test: baseline={10,11,10,10,11}, latest=11
	// z-score≈1.22 < 2.0 → no anomaly
	normalValues := []float64{10, 11, 10, 10, 11, 10, 11}
	result = detector.Detect(normalValues)
	if result.IsAnomaly {
		t.Error("expected no anomaly for normal values")
	}
}
