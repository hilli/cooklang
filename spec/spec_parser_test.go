package spec_test

import (
	"testing"

	"github.com/hilli/cooklang/spec"
)

var (
	specExample = `
version: 7
tests:
  testBasicDirection:
    source: |
      ---
      title: Hot stuff
      ---
      Add a bit of @chilli{5%g}
    result:
      steps:
        -
          - type: text
            value: "Add a bit of"
          - type: ingredient
            name: chilli
            quantity: 5
            units: g
      metadata:
        "title": "Hot stuff"
`
)

func Test_SpecParser(t *testing.T) {
	var res spec.CanonicalTests
	err := spec.ParseSpecData([]byte(specExample), &res)
	if err != nil {
		t.Error(err)
	}
	// t.Log(res.Tests)
	tests := res.Tests
	for _, test := range tests {
		// t.Logf("Source: %+v", test.Source)
		// t.Logf("Result: %+v", test.Result)
		// t.Logf("Steps: %+v", test.Result.Steps)
		if test.Result.Metadata["title"] != "Hot stuff" {
			t.Error("Error passing metadata:", err)
		}
		for _, steps := range test.Result.Steps {
			for i, step := range steps {
				// t.Log("Component", i, step.Type)
				if i == 0 { // The instructions
					if step.Type != "text" {
						t.Error("Expected 'text' type", err)
					}
					if step.Value != "Add a bit of" {
						t.Error("Expected 'Add a bit of', got:", step.Type)
					}
				}
				// Component 1 (The ingredient)
				if i == 1 {
					if step.Type != "ingredient" {
						t.Error("Expected 'ingredient' type", err)
					}
					if step.Name != "chilli" {
						t.Error("Expected 'chilli', got:", step.Name)
					}
					if step.Quantity != "5" {
						t.Error("Expected quantity of '5', got ", step.Quantity)
					}
					if step.Unit != "g" {
						t.Error("Expected unit in 'g', got", step.Unit)
					}
				}
			}
		}
	}
}
