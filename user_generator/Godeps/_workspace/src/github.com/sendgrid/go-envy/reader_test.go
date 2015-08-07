package envy

import (
	"testing"
)

type EnvTest struct {
	description string

	env    []string
	prefix string

	expectedEnv map[string]string
}

func TestEnvParsing(t *testing.T) {
	tests := []EnvTest{
		{
			description: "simple string handling",
			env:         []string{"HOST=localhost", "PORT=9000", "DEBUG=true"},
			expectedEnv: map[string]string{"HOST": "localhost", "PORT": "9000", "DEBUG": "true"},
		},

		{
			description: "prefix values only return the non-prefix keys",
			env:         []string{"KAMTA_HOST=localhost", "KAMTA_PORT=9000", "KAMTA_DEBUG=true"},
			prefix:      "KAMTA_",
			expectedEnv: map[string]string{"HOST": "localhost", "PORT": "9000", "DEBUG": "true"},
		},

		{
			description: "only prefix values are read",
			env: []string{
				"KAMTA_HOST=localhost", "KAMTA_PORT=9000", "KAMTA_DEBUG=true",
				"HOST=remotehost", "PORT=54321", "DEBUG=false",
			},
			prefix:      "KAMTA_",
			expectedEnv: map[string]string{"HOST": "localhost", "PORT": "9000", "DEBUG": "true"},
		},
	}

	for _, test := range tests {
		t.Log("Starting:", test.description)

		// initialize the reader with the prefix and source
		var reader EnvironmentReader
		reader = &OsEnvironmentReader{Prefix: test.prefix, Source: func() []string { return test.env }}

		if reader.GetPrefix() != test.prefix {
			t.Errorf("Expected prefix missing from the EnvironmentReader")
		}

		//
		// Test EnvironmentReader.Read()
		//
		actualEnv := reader.Read()

		if len(actualEnv) != len(test.expectedEnv) {
			t.Errorf("Got an unexpected number of environent variables. Expected %d, got %d", len(test.expectedEnv), len(actualEnv))
		}

		for actualKey, actualValue := range actualEnv {
			expectedValue, ok := test.expectedEnv[actualKey]

			if !ok {
				t.Errorf("Expected value missing from the Read(): %q", actualKey)
			}

			if expectedValue != actualValue {
				t.Errorf("Expected value was incorrect. Got %q, expected %q", actualValue, expectedValue)
			}
		}

		t.Log("Finished:", test.description)
	}
}
