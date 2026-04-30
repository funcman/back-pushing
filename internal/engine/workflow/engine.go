package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/funcman/back-pushing/internal/engine/temporal"
)

type WorkflowEngine struct {
	rules      *RuleSet
	alerts     *AlertManager
	temporal   *temporal.TemporalAnalyzer
	workflows  map[string]func(ctx context.Context, data map[string]any) error
	mu         sync.RWMutex
}

func NewWorkflowEngine(temporal *temporal.TemporalAnalyzer) *WorkflowEngine {
	return &WorkflowEngine{
		rules:     NewRuleSet(),
		alerts:    NewAlertManager(),
		temporal:  temporal,
		workflows: make(map[string]func(ctx context.Context, data map[string]any) error),
	}
}

func (e *WorkflowEngine) AddRule(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules.Add(rule)
}

func (e *WorkflowEngine) RegisterWorkflow(name string, handler func(ctx context.Context, data map[string]any) error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.workflows[name] = handler
}

func (e *WorkflowEngine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error {
	e.mu.RLock()
	triggered := e.rules.Evaluate(data)
	e.mu.RUnlock()

	for _, rule := range triggered {
		action := rule.Action
		switch action.Type {
		case "alert":
			e.alerts.Trigger(
				rule.Name,
				fmt.Sprintf("%v", action.Params["severity"]),
				fmt.Sprintf("Rule %s triggered", rule.Name),
				data,
			)
		case "workflow":
			if workflowName, ok := action.Params["name"].(string); ok {
				e.mu.RLock()
				handler, exists := e.workflows[workflowName]
				e.mu.RUnlock()
				if exists {
					if err := handler(ctx, data); err != nil {
						return err
					}
				}
			}
		case "record_event":
			e.temporal.RecordEvent(ctx, temporal.Event{
				ID:    fmt.Sprintf("event-%d", time.Now().UnixNano()),
				Type:  eventType,
				Actor: fmt.Sprintf("%v", data["actor"]),
				Props: data,
			})
		}
	}

	return nil
}

func (e *WorkflowEngine) GetAlerts() []Alert {
	return e.alerts.GetAlerts()
}
