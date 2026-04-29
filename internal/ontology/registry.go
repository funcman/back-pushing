package ontology

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu          sync.RWMutex
	objectTypes map[string]*ObjectType
	links       map[string]*Link
	paths       map[string]*PathDef
}

func NewRegistry() *Registry {
	return &Registry{
		objectTypes: make(map[string]*ObjectType),
		links:       make(map[string]*Link),
		paths:       make(map[string]*PathDef),
	}
}

func (r *Registry) RegisterOntology(ont *Ontology) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, obj := range ont.ObjectTypes {
		if _, exists := r.objectTypes[name]; exists {
			return fmt.Errorf("object type %s already registered", name)
		}
		objCopy := obj
		r.objectTypes[name] = &objCopy
	}

	for name, link := range ont.Links {
		if _, exists := r.links[name]; exists {
			return fmt.Errorf("link %s already registered", name)
		}
		linkCopy := link
		r.links[name] = &linkCopy
	}

	for name, path := range ont.Paths {
		if _, exists := r.paths[name]; exists {
			return fmt.Errorf("path %s already registered", name)
		}
		pathCopy := path
		r.paths[name] = &pathCopy
	}

	return nil
}

func (r *Registry) GetObjectType(name string) (*ObjectType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	obj, ok := r.objectTypes[name]
	if !ok {
		return nil, fmt.Errorf("object type %s not found", name)
	}
	return obj, nil
}

func (r *Registry) GetLink(name string) (*Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.links[name]
	if !ok {
		return nil, fmt.Errorf("link %s not found", name)
	}
	return link, nil
}

func (r *Registry) GetPath(name string) (*PathDef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path, ok := r.paths[name]
	if !ok {
		return nil, fmt.Errorf("path %s not found", name)
	}
	return path, nil
}

func (r *Registry) ListObjectTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.objectTypes))
	for name := range r.objectTypes {
		names = append(names, name)
	}
	return names
}