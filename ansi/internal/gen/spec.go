package gen

// Spec represents the complete specification for generating ANSI sequences.
type Spec struct {
	Constants map[string]string `yaml:"constants"`
	Sequences []Sequence        `yaml:"sequences"`
	Modes     []Mode            `yaml:"modes"`
}

// Sequence represents a single ANSI escape sequence definition.
type Sequence struct {
	Name      string     `yaml:"name"`
	Aliases   []string   `yaml:"aliases"`
	Doc       string     `yaml:"doc"`
	Params    []Param    `yaml:"params"`
	Format    string     `yaml:"format"`
	Constants []Constant `yaml:"constants"`
}

// Param represents a parameter for a sequence.
type Param struct {
	Name    string      `yaml:"name"`
	Type    string      `yaml:"type"`
	Default interface{} `yaml:"default"`
}

// Constant represents a constant to generate.
type Constant struct {
	Name string        `yaml:"name"`
	Args []interface{} `yaml:"args"`
	Doc  string        `yaml:"doc"`
}

// Mode represents a terminal mode definition.
type Mode struct {
	Name    string   `yaml:"name"`
	Aliases []string `yaml:"aliases"`
	Type    string   `yaml:"type"` // "ansi" or "dec"
	Number  int      `yaml:"number"`
	Doc     string   `yaml:"doc"`
}
