package action

import (
	"testing"

	"github.com/back-pushing/back-pushing/internal/storage/memory"
)

func TestDispatcher_RegisterAndDispatch(t *testing.T) {
	dispatcher := NewDispatcher()
	objStore := memory.NewObjectStore()
	ctx := NewActionContext(objStore)

	dispatcher.Register("escalate_review", EscalateReview)

	input := EscalateReviewInput{
		Reason:   "High risk transaction detected",
		Priority: "high",
		PersonID: "person-123",
	}

	output, err := dispatcher.Dispatch(ctx, "escalate_review", input)
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	result, ok := output.(EscalateReviewOutput)
	if !ok {
		t.Fatalf("expected EscalateReviewOutput, got %T", output)
	}

	if result.CaseID == "" {
		t.Error("expected non-empty CaseID")
	}

	if result.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", result.Status)
	}
}

func TestDispatcher_ActionNotFound(t *testing.T) {
	dispatcher := NewDispatcher()
	objStore := memory.NewObjectStore()
	ctx := NewActionContext(objStore)

	_, err := dispatcher.Dispatch(ctx, "nonexistent_action", nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestDispatcher_ListActions(t *testing.T) {
	dispatcher := NewDispatcher()

	dispatcher.Register("action1", EscalateReview)
	dispatcher.Register("action2", EscalateReview)

	actions := dispatcher.ListActions()
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
}