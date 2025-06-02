package spec

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// ParseSpecFile reads a YAML file and unmarshals it into the provided interface.
// The spec test is defined in the file at the given path.
// File: ../spec/canonical.yaml
func ParseSpecFile(path string, out *CanonicalTests) error {
	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read spec file %s: %w", path, err)
	}

	return ParseSpecData(data, out)
}

func ParseSpecData(data []byte, out *CanonicalTests) error {
	// Unmarshal the YAML content into the provided interface
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal spec: %w", err)
	}
	return nil
}
