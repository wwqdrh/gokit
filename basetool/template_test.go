package basetool

import "testing"

// TestTemplateParse tests the TemplateParse function
func TestTemplateParse(t *testing.T) {
	// Define a table of test cases
	tests := []struct {
		name string      // name of the test case
		str  string      // input template string
		vars interface{} // input variables
		want string      // expected output
	}{
		{
			name: "simple template",
			str:  "Hello, {{.Name}}!",
			vars: map[string]string{"Name": "Alice"},
			want: "Hello, Alice!",
		},
		{
			name: "nested template",
			str:  "{{.Greeting}}, {{.Person.Name}}. You are {{.Person.Age}} years old.",
			vars: map[string]interface{}{
				"Greeting": "Hi",
				"Person": map[string]interface{}{
					"Name": "Bob",
					"Age":  42,
				},
			},
			want: "Hi, Bob. You are 42 years old.",
		},
		{
			name: "invalid template",
			str:  "Hello, {{.Name}}",
			vars: map[string]string{"Name": "Charlie"},
			want: "Hello, Charlie", // expect an empty string
		},
	}

	// Loop over the test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the input
			got := TemplateParse(tc.str, tc.vars)
			// Check if the output matches the expected
			if got != tc.want {
				// Report an error if not
				t.Errorf("TemplateParse(%q, %v) = %q; want %q", tc.str, tc.vars, got, tc.want)
			}
		})
	}
}
