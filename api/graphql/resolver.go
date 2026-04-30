package graphql

import (
    "context"

    "github.com/funcman/back-pushing/internal/storage/memory"
)

type Resolver struct {
    store *memory.ObjectStore
}

func NewResolver(store *memory.ObjectStore) *Resolver {
    return &Resolver{store: store}
}

type Object struct {
    ID   string         `json:"id"`
    Type string         `json:"type"`
    Data map[string]any `json:"data"`
}

func (r *Resolver) Object(ctx context.Context, args struct{ Type, ID string }) (*Object, error) {
    data, err := r.store.Get(ctx, args.Type, args.ID)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: data}, nil
}

func (r *Resolver) Objects(ctx context.Context, args struct{ Type string }) ([]*Object, error) {
    objects, err := r.store.List(ctx, args.Type, nil)
    if err != nil {
        return nil, err
    }

    var result []*Object
    for _, data := range objects {
        result = append(result, &Object{Type: args.Type, Data: data})
    }
    return result, nil
}

func (r *Resolver) CreateObject(ctx context.Context, args struct {
    Type string         `json:"type"`
    ID   string         `json:"id"`
    Data map[string]any `json:"data"`
}) (*Object, error) {
    err := r.store.Create(ctx, args.Type, args.ID, args.Data)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: args.Data}, nil
}

func (r *Resolver) UpdateObject(ctx context.Context, args struct {
    Type string         `json:"type"`
    ID   string         `json:"id"`
    Data map[string]any `json:"data"`
}) (*Object, error) {
    err := r.store.Update(ctx, args.Type, args.ID, args.Data)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: args.Data}, nil
}

func (r *Resolver) DeleteObject(ctx context.Context, args struct{ Type, ID string }) (bool, error) {
    err := r.store.Delete(ctx, args.Type, args.ID)
    return err == nil, err
}