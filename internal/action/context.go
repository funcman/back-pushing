package action

import (
	"context"
	"sync"
	"time"

	"github.com/back-pushing/back-pushing/internal/storage/memory"
)

type ActionContext struct {
	ObjectStore *memory.ObjectStore
	EventLogger *EventLogger
	Workflow    *WorkflowEngine
	AuditLog    *AuditLogger
	mu          sync.RWMutex
	state       map[string]map[string]any
}

func NewActionContext(objStore *memory.ObjectStore) *ActionContext {
	return &ActionContext{
		ObjectStore: objStore,
		EventLogger: NewEventLogger(),
		Workflow:    NewWorkflowEngine(),
		AuditLog:    NewAuditLogger(),
		state:       make(map[string]map[string]any),
	}
}

func (ctx *ActionContext) LogEvent(event *Event) error {
	return ctx.EventLogger.Log(event)
}

func (ctx *ActionContext) TriggerWorkflow(name string, params map[string]any) error {
	return ctx.Workflow.Trigger(name, params)
}

func (ctx *ActionContext) UpdateObject(objType, id string, data map[string]any) error {
	return ctx.ObjectStore.Update(context.Background(), objType, id, data)
}

func (ctx *ActionContext) SetState(key string, value map[string]any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.state[key] == nil {
		ctx.state[key] = make(map[string]any)
	}
	for k, v := range value {
		ctx.state[key][k] = v
	}
}

func (ctx *ActionContext) GetState(key string) map[string]any {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if ctx.state[key] == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range ctx.state[key] {
		result[k] = v
	}
	return result
}

type Event struct {
	Type      string
	Timestamp time.Time
	Payload   any
	Actor     string
}

type EventLogger struct {
	mu     sync.Mutex
	events []Event
}

func NewEventLogger() *EventLogger {
	return &EventLogger{events: make([]Event, 0)}
}

func (l *EventLogger) Log(e *Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	l.events = append(l.events, *e)
	return nil
}

func (l *EventLogger) GetEvents() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Event{}, l.events...)
}

type WorkflowEngine struct {
	mu        sync.Mutex
	workflows map[string]func(map[string]any) error
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{workflows: make(map[string]func(map[string]any) error)}
}

func (w *WorkflowEngine) Register(name string, handler func(map[string]any) error) {
	w.workflows[name] = handler
}

func (w *WorkflowEngine) Trigger(name string, params map[string]any) error {
	w.mu.Lock()
	handler, ok := w.workflows[name]
	w.mu.Unlock()
	if !ok {
		return nil
	}
	return handler(params)
}

type AuditLogger struct {
	mu   sync.Mutex
	logs []AuditEntry
}

type AuditEntry struct {
	Timestamp  time.Time
	Action     string
	ObjectType string
	ObjectID   string
	Actor      string
	Result     string
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{logs: make([]AuditEntry, 0)}
}

func (l *AuditLogger) Log(entry AuditEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	l.logs = append(l.logs, entry)
}