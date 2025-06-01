package spec_test

type CanonicalTests struct {
	Tests map[string]Test `yaml:"tests"`
}

type Test struct {
	// Name   string  `yaml:"name"`
	Source string  `yaml:"source"`
	Result Results `yaml:"result"`
}

type Results struct {
	Steps    [][]Step          `yaml:"steps"`
	Metadata map[string]string `yaml:"metadata"`
}

type Step struct {
	Type     string `yaml:"type"` // ingredient, text, timer or cookware
	Name     string `yaml:"name,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Quantity string `yaml:"quantity,omitempty"`
	Units    string `yaml:"units,omitempty"`
}
