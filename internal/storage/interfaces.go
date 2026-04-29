package storage

import "context"

type ObjectStore interface {
	Create(ctx context.Context, objType string, id string, data map[string]any) error
	Get(ctx context.Context, objType string, id string) (map[string]any, error)
	Update(ctx context.Context, objType string, id string, data map[string]any) error
	Delete(ctx context.Context, objType string, id string) error
	List(ctx context.Context, objType string, filter map[string]any) ([]map[string]any, error)
}

type GraphStore interface {
	AddEdge(ctx context.Context, linkType string, fromID, toID string, props map[string]any) error
	RemoveEdge(ctx context.Context, linkType string, fromID, toID string) error
	GetEdges(ctx context.Context, linkType string, nodeID string) ([]Edge, error)
	Traverse(ctx context.Context, startID string, linkTypes []string, depth int) ([]Path, error)
}

type Edge struct {
	From  string
	To    string
	Props map[string]any
}

type Path struct {
	Nodes []string
	Edges []Edge
}

var (
	ErrNotFound      = &StorageError{msg: "object not found"}
	ErrAlreadyExists = &StorageError{msg: "object already exists"}
)

type StorageError struct {
	msg string
}

func (e *StorageError) Error() string { return e.msg }
