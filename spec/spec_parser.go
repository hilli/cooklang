package spec_test

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseSpecFile reads a YAML file and unmarshals it into the provided interface.
// The spec test is defined in the file at the given path.
// File: ../spec/canonical.yaml
func ParseSpecFile(path string, out any) error {
	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read spec file %s: %w", path, err)
	}

	// Unmarshal the YAML content into the provided interface
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal spec file %s: %w", path, err)
	}
	return nil
}
