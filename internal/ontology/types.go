package ontology

type ObjectType struct {
	Name        string
	Description string
	Type        string // "entity", "event", "concept"
	Properties  map[string]Property
	Links       map[string]LinkDef
	Actions     []ActionDef
	Paths       map[string]PathDef
}

type Property struct {
	Type     string
	Primary  bool
	Indexed  bool
	Unique   bool
	Computed bool
	Source   string
}

type LinkDef struct {
	Target  string
	Through string
	Reverse string
}

type ActionDef struct {
	Name        string
	Description string
	Handler     string
	Args        map[string]string
}

type PathDef struct {
	Description string
	Steps       []string
}

type Ontology struct {
	ObjectTypes    map[string]ObjectType
	Links          map[string]Link
	Paths          map[string]PathDef
	Classification *Classification
}

type Link struct {
	Description string
	Source      string
	Target      string
	Type        string
	Properties  map[string]Property
	Actions     []ActionDef
}

type Classification struct {
	Levels       []string
	DataHandling map[string]DataHandling
	ObjectTags   map[string]ObjectTag
}

type DataHandling struct {
	Description string
	Actions     []string
}

type ObjectTag struct {
	Sensitivity string
	Handling    string
}