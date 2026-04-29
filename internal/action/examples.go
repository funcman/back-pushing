package action

import (
	"fmt"
	"time"
)

type EscalateReviewInput struct {
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
	PersonID string `json:"person_id"`
}

type EscalateReviewOutput struct {
	CaseID     string    `json:"case_id"`
	AssignedTo string    `json:"assigned_to"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

func EscalateReview(ctx *ActionContext, input any) (any, error) {
	in, ok := input.(EscalateReviewInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type")
	}

	caseID := fmt.Sprintf("CASE-%d", time.Now().UnixNano()%1000000)

	event := &Event{
		Type:      "REVIEW_ESCALATION",
		Timestamp: time.Now(),
		Payload:   in,
	}
	ctx.LogEvent(event)

	ctx.Workflow.Register("review_escalation", func(params map[string]any) error {
		return nil
	})
	ctx.TriggerWorkflow("review_escalation", map[string]any{
		"case_id":  caseID,
		"person_id": in.PersonID,
		"priority":  in.Priority,
	})

	ctx.UpdateObject("Person", in.PersonID, map[string]any{
		"review_status":    "escalated",
		"last_escalation": time.Now(),
	})

	return EscalateReviewOutput{
		CaseID:     caseID,
		AssignedTo: "risk_team",
		Status:     "pending",
		Timestamp:  time.Now(),
	}, nil
}