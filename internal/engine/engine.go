package engine

import (
    "context"

    "github.com/funcman/back-pushing/internal/engine/search"
    "github.com/funcman/back-pushing/internal/engine/temporal"
    "github.com/funcman/back-pushing/internal/engine/workflow"
    "github.com/funcman/back-pushing/internal/storage/memory"
)

type Engine struct {
    ObjectStore *memory.ObjectStore
    GraphStore  *memory.GraphStore
    Temporal    *temporal.TemporalAnalyzer
    Search      *search.FilterEngine
    Workflow    *workflow.WorkflowEngine
}

func NewEngine(objStore *memory.ObjectStore, graphStore *memory.GraphStore) *Engine {
    temporalAnalyzer := temporal.NewTemporalAnalyzer()

    return &Engine{
        ObjectStore: objStore,
        GraphStore:  graphStore,
        Temporal:    temporalAnalyzer,
        Search:      search.NewFilterEngine(objStore),
        Workflow:    workflow.NewWorkflowEngine(temporalAnalyzer),
    }
}

func (e *Engine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error {
    return e.Workflow.ProcessEvent(ctx, eventType, data)
}