package workflow

import (
	"fmt"
	"strings"
)

type Rule struct {
	Name        string
	Description string
	Condition   RuleCondition
	Action      RuleAction
	Enabled     bool
}

type RuleCondition struct {
	Field    string
	Operator string
	Value    any
}

type RuleAction struct {
	Type   string
	Params map[string]any
}

const (
	OpEq       = "eq"
	OpNe       = "ne"
	OpGt       = "gt"
	OpLt       = "lt"
	OpGte      = "gte"
	OpLte      = "lte"
	OpContains = "contains"
)

func (r *Rule) Evaluate(data map[string]any) bool {
	if !r.Enabled {
		return false
	}

	value, exists := data[r.Condition.Field]
	if !exists {
		return false
	}

	switch r.Condition.Operator {
	case OpEq:
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", r.Condition.Value)
	case OpNe:
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", r.Condition.Value)
	case OpGt:
		return compareNumeric(value, r.Condition.Value) > 0
	case OpLt:
		return compareNumeric(value, r.Condition.Value) < 0
	case OpGte:
		return compareNumeric(value, r.Condition.Value) >= 0
	case OpLte:
		return compareNumeric(value, r.Condition.Value) <= 0
	case OpContains:
		return strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", r.Condition.Value))
	}

	return false
}

func compareNumeric(a, b any) int {
	af, ok1 := toFloat64(a)
	bf, ok2 := toFloat64(b)
	if ok1 && ok2 {
		if af > bf {
			return 1
		}
		if af < bf {
			return -1
		}
		return 0
	}
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

type RuleSet struct {
	Rules []Rule
}

func NewRuleSet() *RuleSet {
	return &RuleSet{Rules: []Rule{}}
}

func (rs *RuleSet) Add(rule Rule) {
	rs.Rules = append(rs.Rules, rule)
}

func (rs *RuleSet) Evaluate(data map[string]any) []Rule {
	var triggered []Rule
	for _, rule := range rs.Rules {
		if rule.Evaluate(data) {
			triggered = append(triggered, rule)
		}
	}
	return triggered
}
