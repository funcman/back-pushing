package temporal

import (
	"math"
)

type AnomalyDetector struct {
	threshold float64
	window    int
}

func NewAnomalyDetector(threshold float64, windowSize int) *AnomalyDetector {
	return &AnomalyDetector{
		threshold: threshold,
		window:    windowSize,
	}
}

type AnomalyResult struct {
	IsAnomaly bool
	Score     float64
	Message   string
}

func (d *AnomalyDetector) Detect(values []float64) AnomalyResult {
	if len(values) < d.window {
		return AnomalyResult{IsAnomaly: false, Score: 0}
	}

	recent := values[len(values)-d.window:]
	mean := d.mean(recent)
	stddev := d.stddev(recent, mean)

	if len(values) > 0 {
		latest := values[len(values)-1]
		zscore := math.Abs((latest - mean) / stddev)
		if zscore > d.threshold {
			return AnomalyResult{
				IsAnomaly: true,
				Score:     zscore,
				Message:   "Value deviates significantly from recent average",
			}
		}
	}

	return AnomalyResult{IsAnomaly: false, Score: 0}
}

func (d *AnomalyDetector) mean(vals []float64) float64 {
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func (d *AnomalyDetector) stddev(vals []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range vals {
		diff := v - mean
		sum += diff * diff
	}
	variance := sum / float64(len(vals))
	return math.Sqrt(variance)
}
