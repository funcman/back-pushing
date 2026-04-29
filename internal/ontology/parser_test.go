package ontology

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	// Create a temporary directory with test YAML files
	tmpDir := t.TempDir()

	// Create test YAML file with Person object type
	testYAML := `object_types:
  Person:
    description: "A person in the system"
    type: "entity"
    properties:
      id:
        type: "string"
        primary: true
        indexed: true
        unique: true
        computed: false
        source: "system"
      name:
        type: "string"
        primary: false
        indexed: true
        unique: false
        computed: false
        source: "user"
      email:
        type: "string"
        primary: false
        indexed: false
        unique: false
        computed: true
        source: "derived"
    links:
      team:
        target: "Team"
        through: "membership"
        reverse: "members"
    actions:
      - name: "notify"
        description: "Send notification to person"
        handler: "NotifyHandler"
        args:
          channel: "email"
    paths:
      reporting_chain:
        description: "Management reporting chain"
        steps:
          - "manager"
          - "skip_level"
`
	testFile := filepath.Join(tmpDir, "test_ontology.yaml")
	if err := os.WriteFile(testFile, []byte(testYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Parse the ontology
	ontology, err := ParseOntology(tmpDir)
	if err != nil {
		t.Fatalf("ParseOntology() error = %v", err)
	}

	// Verify object types
	if len(ontology.ObjectTypes) != 1 {
		t.Errorf("expected 1 object type, got %d", len(ontology.ObjectTypes))
	}

	// Verify Person object type
	person, ok := ontology.ObjectTypes["Person"]
	if !ok {
		t.Fatal("Person object type not found")
	}

	if person.Description != "A person in the system" {
		t.Errorf("Person.Description = %q, want %q", person.Description, "A person in the system")
	}

	if person.Type != "entity" {
		t.Errorf("Person.Type = %q, want %q", person.Type, "entity")
	}

	// Verify properties
	if len(person.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(person.Properties))
	}

	// Verify primary property (id)
	idProp, ok := person.Properties["id"]
	if !ok {
		t.Fatal("id property not found")
	}
	if idProp.Type != "string" {
		t.Errorf("id.Type = %q, want %q", idProp.Type, "string")
	}
	if !idProp.Primary {
		t.Error("id.Primary should be true")
	}
	if !idProp.Indexed {
		t.Error("id.Indexed should be true")
	}
	if !idProp.Unique {
		t.Error("id.Unique should be true")
	}
	if idProp.Computed {
		t.Error("id.Computed should be false")
	}
	if idProp.Source != "system" {
		t.Errorf("id.Source = %q, want %q", idProp.Source, "system")
	}

	// Verify indexed property (name)
	nameProp, ok := person.Properties["name"]
	if !ok {
		t.Fatal("name property not found")
	}
	if nameProp.Primary {
		t.Error("name.Primary should be false")
	}
	if !nameProp.Indexed {
		t.Error("name.Indexed should be true")
	}
	if nameProp.Unique {
		t.Error("name.Unique should be false")
	}

	// Verify computed property (email)
	emailProp, ok := person.Properties["email"]
	if !ok {
		t.Fatal("email property not found")
	}
	if emailProp.Unique {
		t.Error("email.Unique should be false")
	}
	if !emailProp.Computed {
		t.Error("email.Computed should be true")
	}
	if emailProp.Source != "derived" {
		t.Errorf("email.Source = %q, want %q", emailProp.Source, "derived")
	}

	// Verify links
	if len(person.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(person.Links))
	}

	teamLink, ok := person.Links["team"]
	if !ok {
		t.Fatal("team link not found")
	}
	if teamLink.Target != "Team" {
		t.Errorf("team.Target = %q, want %q", teamLink.Target, "Team")
	}
	if teamLink.Through != "membership" {
		t.Errorf("team.Through = %q, want %q", teamLink.Through, "membership")
	}
	if teamLink.Reverse != "members" {
		t.Errorf("team.Reverse = %q, want %q", teamLink.Reverse, "members")
	}

	// Verify actions
	if len(person.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(person.Actions))
	}
	if person.Actions[0].Name != "notify" {
		t.Errorf("action name = %q, want %q", person.Actions[0].Name, "notify")
	}
	if person.Actions[0].Handler != "NotifyHandler" {
		t.Errorf("action handler = %q, want %q", person.Actions[0].Handler, "NotifyHandler")
	}

	// Verify paths
	if len(person.Paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(person.Paths))
	}
	chainPath, ok := person.Paths["reporting_chain"]
	if !ok {
		t.Fatal("reporting_chain path not found")
	}
	if chainPath.Description != "Management reporting chain" {
		t.Errorf("path description = %q, want %q", chainPath.Description, "Management reporting chain")
	}
	if len(chainPath.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(chainPath.Steps))
	}
}

func TestParseEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	ontology, err := ParseOntology(tmpDir)
	if err != nil {
		t.Fatalf("ParseOntology() error = %v", err)
	}

	if len(ontology.ObjectTypes) != 0 {
		t.Errorf("expected 0 object types, got %d", len(ontology.ObjectTypes))
	}
}

func TestParseNonExistentDirectory(t *testing.T) {
	_, err := ParseOntology("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getString", func(t *testing.T) {
		m := map[string]interface{}{"key": "value", "empty": ""}
		if got := getString(m, "key", "default"); got != "value" {
			t.Errorf("getString() = %q, want %q", got, "value")
		}
		if got := getString(m, "missing", "default"); got != "default" {
			t.Errorf("getString() = %q, want %q", got, "default")
		}
		if got := getString(nil, "key", "default"); got != "default" {
			t.Errorf("getString() = %q, want %q", got, "default")
		}
	})

	t.Run("getBool", func(t *testing.T) {
		m := map[string]interface{}{"key": true, "false": false}
		if got := getBool(m, "key", false); got != true {
			t.Errorf("getBool() = %v, want true", got)
		}
		if got := getBool(m, "missing", true); got != true {
			t.Errorf("getBool() = %v, want true", got)
		}
		if got := getBool(nil, "key", true); got != true {
			t.Errorf("getBool() = %v, want true", got)
		}
	})

	t.Run("getStringSlice", func(t *testing.T) {
		m := map[string]interface{}{"key": []interface{}{"a", "b", "c"}}
		got := getStringSlice(m, "key", nil)
		if len(got) != 3 {
			t.Errorf("getStringSlice() len = %d, want 3", len(got))
		}
		if got := getStringSlice(m, "missing", []string{"default"}); len(got) != 1 || got[0] != "default" {
			t.Errorf("getStringSlice() = %v, want [default]", got)
		}
		if got := getStringSlice(nil, "key", []string{"default"}); len(got) != 1 || got[0] != "default" {
			t.Errorf("getStringSlice() = %v, want [default]", got)
		}
	})
}