package calculator_test

import (
	"testing"

	"github.com/pliliya111/go_final_sprint/internal/calculator"
	"github.com/pliliya111/go_final_sprint/internal/model"
)

func TestPerformOperation(t *testing.T) {
	tests := []struct {
		name        string
		task        *model.Task
		expected    interface{}
		expectError bool
	}{
		{
			name: "Addition",
			task: &model.Task{
				Arg1:      float64(2),
				Arg2:      float64(3),
				Operation: "+",
			},
			expected:    float64(5),
			expectError: false,
		},
		{
			name: "Subtraction",
			task: &model.Task{
				Arg1:      float64(5),
				Arg2:      float64(3),
				Operation: "-",
			},
			expected:    float64(2),
			expectError: false,
		},
		{
			name: "Multiplication",
			task: &model.Task{
				Arg1:      float64(4),
				Arg2:      float64(3),
				Operation: "*",
			},
			expected:    float64(12),
			expectError: false,
		},
		{
			name: "Division",
			task: &model.Task{
				Arg1:      float64(10),
				Arg2:      float64(2),
				Operation: "/",
			},
			expected:    float64(5),
			expectError: false,
		},
		{
			name: "Division by zero",
			task: &model.Task{
				Arg1:      float64(10),
				Arg2:      float64(0),
				Operation: "/",
			},
			expected:    "division by zero",
			expectError: true,
		},
		{
			name: "Invalid operation",
			task: &model.Task{
				Arg1:      float64(10),
				Arg2:      float64(2),
				Operation: "invalid",
			},
			expected:    "unknown operation: invalid",
			expectError: true,
		},
		{
			name: "Invalid argument (arg1)",
			task: &model.Task{
				Arg1:      "not_a_number",
				Arg2:      float64(2),
				Operation: "+",
			},
			expected:    "invalid argument: arg1=not_a_number (cannot convert to float64)",
			expectError: true,
		},
		{
			name: "Invalid argument (arg2)",
			task: &model.Task{
				Arg1:      float64(2),
				Arg2:      "not_a_number",
				Operation: "+",
			},
			expected:    "invalid argument: arg2=not_a_number (cannot convert to float64)",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.PerformOperation(tt.task)

			if tt.expectError {
				if result == nil {
					t.Errorf("Expected an error, but got nil")
				} else if result.(error).Error() != tt.expected {
					t.Errorf("Expected error: %v, got: %v", tt.expected, result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected: %v, got: %v", tt.expected, result)
				}
			}
		})
	}
}
