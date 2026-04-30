package temporal

import (
	"time"
)

type WindowType string

const (
	WindowTumbling WindowType = "tumbling"
	WindowSliding  WindowType = "sliding"
)

type Window struct {
	Size       time.Duration
	Slide      time.Duration
	WindowType WindowType
}

type Aggregation struct {
	Count int64
	Sum   float64
	Avg   float64
	Min   float64
	Max   float64
}

func NewTumblingWindow(size time.Duration) *Window {
	return &Window{
		Size:       size,
		Slide:      size,
		WindowType: WindowTumbling,
	}
}

func NewSlidingWindow(size, slide time.Duration) *Window {
	return &Window{
		Size:       size,
		Slide:      slide,
		WindowType: WindowSliding,
	}
}

func (a *Aggregation) Add(t time.Time, value float64) {
	a.Count++
	a.Sum += value
	a.Avg = a.Sum / float64(a.Count)
	if value < a.Min || a.Min == 0 {
		a.Min = value
	}
	if value > a.Max {
		a.Max = value
	}
}

func (w *Window) Aggregate(events []Event, valueField string) *Aggregation {
	agg := &Aggregation{}
	for _, e := range events {
		if v, ok := e.Props[valueField].(float64); ok {
			agg.Add(e.Timestamp, v)
		}
	}
	return agg
}
