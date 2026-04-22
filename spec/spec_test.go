package spec_test

// Are we parsing the spec tests of Cooklang

import (
	"os"
	"reflect"
	"testing"

	"github.com/hilli/cooklang/parser"
	spec_test "github.com/hilli/cooklang/spec"
)

func Test_Spec(t *testing.T) {
	for _, specFile := range []string{"canonical.yaml", "extended.yaml", "canonical_extensions.yaml"} {
		t.Run(specFile, func(t *testing.T) {
			// Don't know if we are going to keep all specs forever, so lets not break the tests if a spec file is missing or empty.
			if fileInfo, err := os.Stat(specFile); os.IsNotExist(err) || fileInfo.Size() == 0 {
				t.Skip("Skipping test for spec file", specFile, "because it does not exist or is empty")
			}

			// Fresh specification per file to avoid test bleed between spec files
			var specification spec_test.CanonicalTests
			err := spec_test.ParseSpecFile(specFile, &specification)
			if err != nil {
				t.Fatalf("Failed to parse spec file %s: %v", specFile, err)
			}

			p := parser.New()
			// Enable extended mode for extended.yaml and canonical_extensions.yaml spec files
			if specFile == "extended.yaml" || specFile == "canonical_extensions.yaml" {
				p.ExtendedMode = true
			}
			for testName, spec := range specification.Tests {
				t.Run(testName, func(t *testing.T) {
					source := spec.Source
					recipe, err := p.ParseString(source)
					if err != nil {
						t.Error(err)
					}

					if len(recipe.Steps) != len(spec.Result.Steps) {
						t.Errorf("step count mismatch: got %d, want %d", len(recipe.Steps), len(spec.Result.Steps))
						t.SkipNow() // skip the rest of the test if steps don't match
					}
					for is, specstep := range spec.Result.Steps {
						recipeComponent := recipe.Steps[is].Components
						if !reflect.DeepEqual(recipeComponent, specstep) {
							t.Errorf("step %d mismatch:\nWant: %#v\nGot : %#v", is, specstep, recipeComponent)
						}
					}

					// Compare metadata if the spec defines any
					if len(spec.Result.Metadata) > 0 {
						if !reflect.DeepEqual(recipe.Metadata, spec.Result.Metadata) {
							t.Errorf("metadata mismatch:\nWant: %#v\nGot : %#v", spec.Result.Metadata, recipe.Metadata)
						}
					}
				})
			}
		})
	}

}
