package npm_test

import (
	"encoding/json"
	"testing"

	"github.com/howmanysmall/npm-registry-mcp/src/npm"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"unmarshals integer", `{"value": 1131}`, 1131, false},
		{"unmarshals string number", `{"value": "1131"}`, 1131, false},
		{"unmarshals zero", `{"value": 0}`, 0, false},
		{"unmarshals zero string", `{"value": "0"}`, 0, false},
		{"unmarshals empty string as zero", `{"value": ""}`, 0, false},
		{"errors on invalid string", `{"value": "not a number"}`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result struct {
				Value npm.FlexInt `json:"value"`
			}

			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			if int(result.Value) != tt.expected {
				t.Errorf("got %d, want %d", result.Value, tt.expected)
			}
		})
	}
}

func TestFlexLicense_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"unmarshals string", `"MIT"`, "MIT", false},
		{"unmarshals object", `{"type": "MIT", "url": "http://example.com"}`, "MIT", false},
		{"unmarshals null", `null`, "", false},
		{"unmarshals empty string", `""`, "", false},
		{"errors on invalid type", `123`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var l npm.FlexLicense

			err := json.Unmarshal([]byte(tt.input), &l)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			if string(l) != tt.expected {
				t.Errorf("got %q, want %q", l, tt.expected)
			}
		})
	}
}

func TestPackageVersion_UnmarshalJSON_WithObjectLicense(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"name": "gray-matter",
		"version": "1.1.0",
		"license": {
			"type": "MIT",
			"url": "https://github.com/assemble/gray-matter/blob/master/LICENSE-MIT"
		}
	}`

	var result npm.PackageVersion

	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if string(result.License) != "MIT" {
		t.Errorf("got license %q, want %q", result.License, "MIT")
	}
}

func TestSearchObject_UnmarshalJSON_WithStringDependents(t *testing.T) {
	t.Parallel()

	// This tests the real-world scenario from NPM API
	jsonData := `{"package":{"name":"jsonc-parser","version":"3.3.1","description":"test","keywords":[],"license":"MIT","date":"2024-06-24T21:12:45.445Z","publisher":{"username":"vscode-bot","email":"vscode-bot-npm@microsoft.com"},"maintainers":[],"links":{"npm":"https://www.npmjs.com/package/jsonc-parser","homepage":"https://github.com/microsoft/node-jsonc-parser#readme","repository":"git+https://github.com/microsoft/node-jsonc-parser.git","bugs":"https://github.com/microsoft/node-jsonc-parser/issues"}},"score":{"final":183.08606,"detail":{"quality":1,"popularity":1,"maintenance":1}},"searchScore":183.08606,"downloads":{"weekly":13660649,"monthly":91955409},"dependents":"1131"}`

	var result npm.SearchObject

	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if int(result.Dependents) != 1131 {
		t.Errorf("got %d dependents, want 1131", result.Dependents)
	}
}
