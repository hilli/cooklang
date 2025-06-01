package parser

// Are we parsing the spec tests of Cooklang

import (
	"testing"

	"github.com/hilli/cooklang/spec_test"
)

func Test_Spec(t *testing.T) {
	spec_test.ParseSpecFile("../spec/canonical.yaml")
}
