package spec_test

// Are we parsing the spec tests of Cooklang

import (
	"reflect"
	"testing"

	"github.com/hilli/cooklang/parser"
	spec_test "github.com/hilli/cooklang/spec"
)

func Test_Spec(t *testing.T) {
	var specification spec_test.CanonicalTests
	err := spec_test.ParseSpecFile("../spec/canonical.yaml", &specification)
	if err != nil {
		t.Fatalf("Failed to parse spec file: %v", err)
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
			}

			for is, specstep := range spec.Result.Steps {
				recipeComponent := recipe.Steps[is].Components
				if !reflect.DeepEqual(recipeComponent, specstep) {
					t.Errorf("Error: %s\nWant: %#v\nGot : %#v", err, specstep, recipeComponent)
				}
			}

			// for i, step := range recipe.Steps {
			// 	t.Log("Step", i+1)
			// 	for _, comp := range step.Components {
			// 		t.Logf("\t%+v", comp)
			// 	}
			// }
		})
	}
}
