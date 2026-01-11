package spec

import "github.com/hilli/cooklang/parser"

type CanonicalTests struct {
	Tests map[string]Test `yaml:"tests"`
}

type Test struct {
	Source string  `yaml:"source"`
	Result Results `yaml:"result"`
}

type Results struct {
	Steps    [][]parser.Component `yaml:"steps"`
	Metadata parser.Metadata      `yaml:"metadata"`
}
