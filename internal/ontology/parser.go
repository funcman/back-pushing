package ontology

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Parse reads all .yaml files in the given directory and returns an Ontology
func Parse(dir string) (*Ontology, error) {
	ontology := &Ontology{
		ObjectTypes: make(map[string]ObjectType),
		Links:       make(map[string]Link),
		Paths:       make(map[string]PathDef),
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		var raw rawOntology
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		// Parse object types
		for name, rawObjType := range raw.ObjectTypes {
			objType := ObjectType{
				Name:        name,
				Description: rawObjType.Description,
				Type:        rawObjType.Type,
				Properties:  make(map[string]Property),
				Links:       make(map[string]LinkDef),
				Actions:     make([]ActionDef, 0),
				Paths:       make(map[string]PathDef),
			}

			// Parse properties
			for propName, rawProp := range rawObjType.Properties {
				objType.Properties[propName] = Property{
					Type:     rawProp.Type,
					Primary:  rawProp.Primary,
					Indexed:  rawProp.Indexed,
					Unique:   rawProp.Unique,
					Computed: rawProp.Computed,
					Source:   rawProp.Source,
				}
			}

			// Parse links
			for linkName, rawLink := range rawObjType.Links {
				objType.Links[linkName] = LinkDef{
					Target:  rawLink.Target,
					Through: rawLink.Through,
					Reverse: rawLink.Reverse,
				}
			}

			// Parse actions
			for _, rawAction := range rawObjType.Actions {
				action := ActionDef{
					Name:        rawAction.Name,
					Description: rawAction.Description,
					Handler:     rawAction.Handler,
					Args:        rawAction.Args,
				}
				objType.Actions = append(objType.Actions, action)
			}

			// Parse paths
			for pathName, rawPath := range rawObjType.Paths {
				objType.Paths[pathName] = PathDef{
					Description: rawPath.Description,
					Steps:       rawPath.Steps,
				}
			}

			ontology.ObjectTypes[name] = objType
		}

		// Parse links at ontology level
		for name, rawLink := range raw.Links {
			link := Link{
				Description: rawLink.Description,
				Source:      rawLink.Source,
				Target:      rawLink.Target,
				Type:        rawLink.Type,
				Properties:  make(map[string]Property),
				Actions:     make([]ActionDef, 0),
			}

			for propName, rawProp := range rawLink.Properties {
				link.Properties[propName] = Property{
					Type:     rawProp.Type,
					Primary:  rawProp.Primary,
					Indexed:  rawProp.Indexed,
					Unique:   rawProp.Unique,
					Computed: rawProp.Computed,
					Source:   rawProp.Source,
				}
			}

			for _, rawAction := range rawLink.Actions {
				link.Actions = append(link.Actions, ActionDef{
					Name:        rawAction.Name,
					Description: rawAction.Description,
					Handler:     rawAction.Handler,
					Args:        rawAction.Args,
				})
			}

			ontology.Links[name] = link
		}

		// Parse paths at ontology level
		for name, rawPath := range raw.Paths {
			ontology.Paths[name] = PathDef{
				Description: rawPath.Description,
				Steps:       rawPath.Steps,
			}
		}

		// Parse classification
		if raw.Classification != nil {
			ontology.Classification = &Classification{
				Levels:       raw.Classification.Levels,
				DataHandling: make(map[string]DataHandling),
				ObjectTags:   make(map[string]ObjectTag),
			}

			for name, dh := range raw.Classification.DataHandling {
				ontology.Classification.DataHandling[name] = DataHandling{
					Description: dh.Description,
					Actions:     dh.Actions,
				}
			}

			for name, ot := range raw.Classification.ObjectTags {
				ontology.Classification.ObjectTags[name] = ObjectTag{
					Sensitivity: ot.Sensitivity,
					Handling:    ot.Handling,
				}
			}
		}
	}

	return ontology, nil
}

// Helper functions

func getString(m map[string]interface{}, key string, defaultVal string) string {
	if m == nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if m == nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getStringSlice(m map[string]interface{}, key string, defaultVal []string) []string {
	if m == nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		if s, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(s))
			for _, item := range s {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return defaultVal
}

// rawOntology represents the raw YAML structure
type rawOntology struct {
	ObjectTypes    map[string]rawObjectType    `yaml:"object_types"`
	Links          map[string]rawLink          `yaml:"links"`
	Paths          map[string]rawPathDef       `yaml:"paths"`
	Classification *rawClassification          `yaml:"classification"`
}

type rawObjectType struct {
	Description string
	Type        string                      `yaml:"type"`
	Properties  map[string]rawProperty      `yaml:"properties"`
	Links       map[string]rawLinkDef       `yaml:"links"`
	Actions     []rawActionDef              `yaml:"actions"`
	Paths       map[string]rawPathDef       `yaml:"paths"`
}

type rawProperty struct {
	Type     string `yaml:"type"`
	Primary  bool   `yaml:"primary"`
	Indexed  bool   `yaml:"indexed"`
	Unique   bool   `yaml:"unique"`
	Computed bool   `yaml:"computed"`
	Source   string `yaml:"source"`
}

type rawLinkDef struct {
	Target  string `yaml:"target"`
	Through string `yaml:"through"`
	Reverse string `yaml:"reverse"`
}

type rawActionDef struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Handler     string            `yaml:"handler"`
	Args        map[string]string `yaml:"args"`
}

type rawPathDef struct {
	Description string   `yaml:"description"`
	Steps       []string `yaml:"steps"`
}

type rawLink struct {
	Description string           `yaml:"description"`
	Source      string           `yaml:"source"`
	Target      string           `yaml:"target"`
	Type        string           `yaml:"type"`
	Properties  map[string]rawProperty `yaml:"properties"`
	Actions     []rawActionDef   `yaml:"actions"`
}

type rawClassification struct {
	Levels       []string                     `yaml:"levels"`
	DataHandling map[string]rawDataHandling   `yaml:"data_handling"`
	ObjectTags   map[string]rawObjectTag     `yaml:"object_tags"`
}

type rawDataHandling struct {
	Description string   `yaml:"description"`
	Actions     []string `yaml:"actions"`
}

type rawObjectTag struct {
	Sensitivity string `yaml:"sensitivity"`
	Handling    string `yaml:"handling"`
}