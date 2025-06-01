package spec_test

import (
	"testing"
)

func TestParseSpecFile(t *testing.T) {
	// type Spec any
	type Spec CanonicalTests
	var spec Spec
	err := ParseSpecFile("../spec/canonical.yaml", &spec)
	if err != nil {
		t.Fatalf("ParseSpecFile failed: %v", err)
	}
	// fmt.Printf("Parsed spec: %+v\n", spec)
	if spec.Tests != "tests" {
		t.Errorf("Expected spec name to be 'tests', got '%s'", spec.Tests)
	}
	// Additional checks can be added here based on the spec structure
}
