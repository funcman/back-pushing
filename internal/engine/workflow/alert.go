package workflow

import (
	"fmt"
	"sync"
	"time"
)

type Alert struct {
	ID        string
	RuleName  string
	Severity  string
	Message   string
	Timestamp time.Time
	Data      map[string]any
}

type AlertHandler func(alert Alert)

type AlertManager struct {
	mu      sync.RWMutex
	alerts  []Alert
	handlers []AlertHandler
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:   []Alert{},
		handlers: []AlertHandler{},
	}
}

func (m *AlertManager) RegisterHandler(handler AlertHandler) {
	m.handlers = append(m.handlers, handler)
}

func (m *AlertManager) Trigger(ruleName, severity, message string, data map[string]any) {
	alert := Alert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		RuleName:  ruleName,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	m.mu.Unlock()

	for _, handler := range m.handlers {
		handler(alert)
	}
}

func (m *AlertManager) GetAlerts() []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Alert{}, m.alerts...)
}

func (m *AlertManager) ClearAlerts() {
	m.mu.Lock()
	m.alerts = []Alert{}
	m.mu.Unlock()
}
