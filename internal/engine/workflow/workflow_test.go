package workflow

import (
	"context"
	"testing"

	"github.com/funcman/back-pushing/internal/engine/temporal"
)

func TestRuleSet_Evaluate(t *testing.T) {
	rs := NewRuleSet()

	rs.Add(Rule{
		Name:        "high_value_transaction",
		Description: "Alert on high value transactions",
		Condition:   RuleCondition{Field: "amount", Operator: OpGt, Value: float64(10000)},
		Action:      RuleAction{Type: "alert", Params: map[string]any{"severity": "high"}},
		Enabled:     true,
	})

	data := map[string]any{"amount": float64(15000), "actor": "acc1"}

	triggered := rs.Evaluate(data)
	if len(triggered) != 1 {
		t.Errorf("expected 1 triggered rule, got %d", len(triggered))
	}

	lowValueData := map[string]any{"amount": float64(100), "actor": "acc1"}
	triggered = rs.Evaluate(lowValueData)
	if len(triggered) != 0 {
		t.Errorf("expected 0 triggered rules, got %d", len(triggered))
	}
}

func TestWorkflowEngine_ProcessEvent(t *testing.T) {
	temporalAnalyzer := temporal.NewTemporalAnalyzer()
	engine := NewWorkflowEngine(temporalAnalyzer)
	ctx := context.Background()

	engine.AddRule(Rule{
		Name:        "high_value",
		Condition:   RuleCondition{Field: "amount", Operator: OpGt, Value: float64(1000)},
		Action:      RuleAction{Type: "alert", Params: map[string]any{"severity": "warning"}},
		Enabled:     true,
	})

	var alertReceived bool
	engine.alerts.RegisterHandler(func(alert Alert) {
		alertReceived = true
	})

	engine.ProcessEvent(ctx, "Transaction", map[string]any{
		"amount": float64(2000),
		"actor":  "acc1",
	})

	if !alertReceived {
		t.Error("expected alert to be received")
	}
}
