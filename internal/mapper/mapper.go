// internal/mapper/mapper.go
package mapper

import (
	"context"
	"fmt"
	"time"

	"github.com/funcman/back-pushing/internal/adapter"
)

type Object struct {
	ID    string
	Type  string
	Data  map[string]any
	Links []Link
}

type Link struct {
	FromID   string
	ToID     string
	LinkType string
	Props    map[string]any
}

type Mapper struct {
	config *MappingConfig
}

func New(cfg *MappingConfig) *Mapper {
	return &Mapper{config: cfg}
}

func (m *Mapper) Map(ctx context.Context, source adapter.DataSource) ([]Object, error) {
	rows, err := source.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	var objects []Object
	for _, row := range rows {
		obj := Object{
			Type: m.config.Target.ObjectType,
			Data: make(map[string]any),
		}

		if id, ok := row["id"]; ok {
			obj.ID = fmt.Sprintf("%v", id)
		}

		for _, f := range m.config.Fields {
			val := row[f.Source]
			converted, err := m.convert(val, f.Type)
			if err != nil {
				return nil, fmt.Errorf("convert field %s: %w", f.Source, err)
			}
			obj.Data[f.Target] = converted

			if f.Link != "" && val != nil {
				obj.Links = append(obj.Links, Link{
					FromID:   obj.ID,
					ToID:     fmt.Sprintf("%v", val),
					LinkType: f.Link,
				})
			}
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

func (m *Mapper) convert(val any, typ string) (any, error) {
	if val == nil {
		return nil, nil
	}

	switch typ {
	case "string":
		return fmt.Sprintf("%v", val), nil
	case "int":
		switch v := val.(type) {
		case float64:
			return int(v), nil
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i, nil
		}
	case "float":
		switch v := val.(type) {
		case float64:
			return v, nil
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f, nil
		}
	case "datetime":
		switch v := val.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, nil
			}
			return v, nil
		}
	}

	return val, nil
}