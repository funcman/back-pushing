package ontology

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Run("register ontology with Person and knows link", func(t *testing.T) {
		reg := NewRegistry()

		ont := &Ontology{
			ObjectTypes: map[string]ObjectType{
				"Person": {
					Name:        "Person",
					Description: "A person in the system",
					Type:        "entity",
					Properties: map[string]Property{
						"name": {Type: "string", Primary: true},
					},
				},
			},
			Links: map[string]Link{
				"knows": {
					Description: "A person who knows another person",
					Source:      "Person",
					Target:      "Person",
					Type:        "directed",
				},
			},
		}

		if err := reg.RegisterOntology(ont); err != nil {
			t.Fatalf("failed to register ontology: %v", err)
		}
	})

	t.Run("GetObjectType returns correct type", func(t *testing.T) {
		reg := NewRegistry()

		ont := &Ontology{
			ObjectTypes: map[string]ObjectType{
				"Person": {
					Name:        "Person",
					Description: "A person in the system",
					Type:        "entity",
				},
			},
		}

		if err := reg.RegisterOntology(ont); err != nil {
			t.Fatalf("failed to register ontology: %v", err)
		}

		obj, err := reg.GetObjectType("Person")
		if err != nil {
			t.Fatalf("failed to get object type: %v", err)
		}
		if obj.Name != "Person" {
			t.Errorf("expected Person, got %s", obj.Name)
		}
		if obj.Type != "entity" {
			t.Errorf("expected entity, got %s", obj.Type)
		}
	})

	t.Run("GetLink returns correct link", func(t *testing.T) {
		reg := NewRegistry()

		ont := &Ontology{
			Links: map[string]Link{
				"knows": {
					Description: "A person who knows another person",
					Source:      "Person",
					Target:      "Person",
					Type:        "directed",
				},
			},
		}

		if err := reg.RegisterOntology(ont); err != nil {
			t.Fatalf("failed to register ontology: %v", err)
		}

		link, err := reg.GetLink("knows")
		if err != nil {
			t.Fatalf("failed to get link: %v", err)
		}
		if link.Source != "Person" {
			t.Errorf("expected Person, got %s", link.Source)
		}
		if link.Target != "Person" {
			t.Errorf("expected Person, got %s", link.Target)
		}
		if link.Type != "directed" {
			t.Errorf("expected directed, got %s", link.Type)
		}
	})

	t.Run("error for non-existent type", func(t *testing.T) {
		reg := NewRegistry()

		_, err := reg.GetObjectType("NonExistent")
		if err == nil {
			t.Error("expected error for non-existent type")
		}
	})
}