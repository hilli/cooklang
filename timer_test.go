package cooklang

import (
	"testing"
)

func TestTimerRenderDisplay(t *testing.T) {
	tests := []struct {
		name     string
		timer    Timer
		expected string
	}{
		{
			name: "duration with unit",
			timer: Timer{
				Duration: "10",
				Unit:     "minutes",
			},
			expected: "10 minutes",
		},
		{
			name: "duration range with unit",
			timer: Timer{
				Duration: "10-15",
				Unit:     "seconds",
			},
			expected: "10-15 seconds",
		},
		{
			name: "duration without unit",
			timer: Timer{
				Duration: "5",
			},
			expected: "5",
		},
		{
			name: "named timer without duration falls back to name",
			timer: Timer{
				Name: "resting",
			},
			expected: "resting",
		},
		{
			name: "duration takes precedence over name",
			timer: Timer{
				Duration: "10",
				Unit:     "minutes",
				Name:     "boiling",
			},
			expected: "10 minutes",
		},
		{
			name:     "empty timer returns empty string",
			timer:    Timer{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timer.RenderDisplay()
			if result != tt.expected {
				t.Errorf("RenderDisplay() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTimerParsingWithUnit(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedDuration string
		expectedUnit     string
		expectedDisplay  string
	}{
		{
			name:             "simple timer with minutes",
			input:            "Boil for ~{10%minutes}.",
			expectedDuration: "10",
			expectedUnit:     "minutes",
			expectedDisplay:  "10 minutes",
		},
		{
			name:             "timer with seconds",
			input:            "Stir for ~{30%seconds}.",
			expectedDuration: "30",
			expectedUnit:     "seconds",
			expectedDisplay:  "30 seconds",
		},
		{
			name:             "timer with range and unit",
			input:            "Mix for ~{10-15%seconds}.",
			expectedDuration: "10-15",
			expectedUnit:     "seconds",
			expectedDisplay:  "10-15 seconds",
		},
		{
			name:             "timer with hours",
			input:            "Bake for ~{2%hours}.",
			expectedDuration: "2",
			expectedUnit:     "hours",
			expectedDisplay:  "2 hours",
		},
		{
			name:             "timer without unit",
			input:            "Wait ~{5}.",
			expectedDuration: "5",
			expectedUnit:     "",
			expectedDisplay:  "5",
		},
		{
			name:             "named timer with unit",
			input:            "Let it ~rest{10%minutes}.",
			expectedDuration: "10",
			expectedUnit:     "minutes",
			expectedDisplay:  "10 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse recipe: %v", err)
			}

			// Find the timer in the parsed recipe
			var foundTimer *Timer
			step := recipe.FirstStep
			for step != nil && foundTimer == nil {
				component := step.FirstComponent
				for component != nil {
					if timer, ok := component.(*Timer); ok {
						foundTimer = timer
						break
					}
					component = component.GetNext()
				}
				step = step.NextStep
			}

			if foundTimer == nil {
				t.Fatal("No timer found in parsed recipe")
			}

			if foundTimer.Duration != tt.expectedDuration {
				t.Errorf("Duration = %q, want %q", foundTimer.Duration, tt.expectedDuration)
			}

			if foundTimer.Unit != tt.expectedUnit {
				t.Errorf("Unit = %q, want %q", foundTimer.Unit, tt.expectedUnit)
			}

			display := foundTimer.RenderDisplay()
			if display != tt.expectedDisplay {
				t.Errorf("RenderDisplay() = %q, want %q", display, tt.expectedDisplay)
			}
		})
	}
}

func TestTimerInRecipeSteps(t *testing.T) {
	// Test that timers are correctly rendered in recipe step text
	recipe, err := ParseString("Stir for ~{10-15%seconds}.")
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	// Get the step instruction text
	step := recipe.FirstStep
	if step == nil {
		t.Fatal("No steps found in recipe")
	}

	// Build the instruction text from components
	var instructionText string
	component := step.FirstComponent
	for component != nil {
		switch c := component.(type) {
		case *Timer:
			instructionText += c.RenderDisplay()
		case *Instruction:
			instructionText += c.Text
		}
		component = component.GetNext()
	}

	// The instruction should contain "10-15 seconds"
	expectedSubstring := "10-15 seconds"
	if !contains(instructionText, expectedSubstring) {
		t.Errorf("Instruction text %q should contain %q", instructionText, expectedSubstring)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
