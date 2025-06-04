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
	var specification spec_test.CanonicalTests

	for _, specFile := range []string{"canonical.yaml", "extended.yaml"} {
		t.Run(specFile, func(t *testing.T) {
			// Don't know if we are going to keep all specs forever, so lets not break the tests if a spec file is missing or empty.
			if fileInfo, err := os.Stat(specFile); os.IsNotExist(err) || fileInfo.Size() == 0 {
				t.Skip("Skipping test for spec file", specFile, "because it does not exist or is empty")
			}
			err := spec_test.ParseSpecFile(specFile, &specification)
			if err != nil {
				t.Fatalf("Failed to parse spec file %s: %v", specFile, err)
			}

			p := parser.New()
			for testName, spec := range specification.Tests {
				t.Run(testName, func(t *testing.T) {
					source := spec.Source
					recipe, err := p.ParseString(source)
					if err != nil {
						t.Error(err)
					}

					if len(recipe.Steps) != len(spec.Result.Steps) {
						t.Error("parsed recipe does not have as many steps as spec", err)
						t.SkipNow() // skip the rest of the test if steps don't match
					}
					for is, specstep := range spec.Result.Steps {
						recipeComponent := recipe.Steps[is].Components
						if !reflect.DeepEqual(recipeComponent, specstep) {
							t.Errorf("Error: %s\nWant: %#v\nGot : %#v", err, specstep, recipeComponent)
						}
					}
				})
			}
		})
	}

}
