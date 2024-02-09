package basetool

import (
	"reflect"
	"testing"
)

// TestIsBlank tests the IsBlank function
func TestIsBlank(t *testing.T) {
	// Define a table of test cases
	tests := []struct {
		name  string      // name of the test case
		value interface{} // input value
		want  bool        // expected output
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
		{"false bool", false, true},
		{"true bool", true, false},
		// {"nil interface", nil, true}, // cant use nil
		{"non-nil interface", 1, false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", new(int), false},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"custom struct", struct{ x int }{0}, false}, // struct values are never blank
	}

	// Loop over the test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the input
			got := IsBlank(reflect.ValueOf(tc.value))
			// Check if the output matches the expected
			if got != tc.want {
				// Report an error if not
				t.Errorf("IsBlank(%v) = %v; want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestStr2Value(t *testing.T) {
	// Define a table of test cases
	tests := []struct {
		name string      // name of the test case
		val  string      // input value
		mode string      // input mode
		want interface{} // expected output
	}{
		{"string to string", "hello", "string", "hello"},
		{"string to []string", "a,b,c", "[]string", []string{"a", "b", "c"}},
		{"string to int", "42", "int", 42},
		{"string to []int", "1,2,3", "[]int", []int{1, 2, 3}},
		{"string to bool", "true", "bool", true},
		{"string to []bool", "true,false,true", "[]bool", []bool{true, false, true}},
		{"string to float", "3.14", "float", 3.14},
		{"string to []float", "1.1,2.2,3.3", "[]float", []float64{1.1, 2.2, 3.3}},
		{"invalid mode", "hello", "invalid", nil},
		{"invalid value", "hello", "int", nil},
	}

	// Loop over the test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the input
			got, err := Str2Value(tc.val, tc.mode)
			// Check if the output matches the expected
			if !reflect.DeepEqual(got, tc.want) {
				// Report an error if not
				t.Errorf("Str2Value(%q, %q) = %v, %v; want %v, nil", tc.val, tc.mode, got, err, tc.want)
			}
		})
	}
}
