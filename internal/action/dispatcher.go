package action

import "fmt"

type Handler func(ctx *ActionContext, input any) (output any, err error)

type Dispatcher struct {
	mu map[string]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{mu: make(map[string]Handler)}
}

func (d *Dispatcher) Register(name string, handler Handler) {
	d.mu[name] = handler
}

func (d *Dispatcher) Dispatch(ctx *ActionContext, name string, input any) (any, error) {
	handler, ok := d.mu[name]
	if !ok {
		return nil, fmt.Errorf("action %s not found", name)
	}
	output, err := handler(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("action %s failed: %w", name, err)
	}
	return output, nil
}

func (d *Dispatcher) ListActions() []string {
	actions := make([]string, 0, len(d.mu))
	for name := range d.mu {
		actions = append(actions, name)
	}
	return actions
}