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

func TestFlexEngines_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]string
		wantErr  bool
	}{
		{"unmarshals map", `{"node": ">=14"}`, map[string]string{"node": ">=14"}, false},
		{"unmarshals array", `["node >= 0.10.0", "npm >= 1.2.0"]`, map[string]string{"node >= 0.10.0": "*", "npm >= 1.2.0": "*"}, false},
		{"unmarshals empty map", `{}`, map[string]string{}, false},
		{"unmarshals empty array", `[]`, map[string]string{}, false},
		{"errors on invalid type", `123`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var e npm.FlexEngines

			err := json.Unmarshal([]byte(tt.input), &e)

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

			if len(e) != len(tt.expected) {
				t.Errorf("got length %d, want %d", len(e), len(tt.expected))
			}

			for k, v := range tt.expected {
				if e[k] != v {
					t.Errorf("key %q: got %q, want %q", k, e[k], v)
				}
			}
		})
	}
}

func TestRepository_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"unmarshals object", `{"type": "git", "url": "https://github.com/foo/bar"}`, "https://github.com/foo/bar"},
		{"unmarshals string github", `"github:foo/bar"`, "https://github.com/foo/bar"},
		{"unmarshals string url", `"https://github.com/foo/bar"`, "https://github.com/foo/bar"},
		{"unmarshals empty string", `""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r npm.Repository

			err := json.Unmarshal([]byte(tt.input), &r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if r.URL != tt.expected {
				t.Errorf("got %q, want %q", r.URL, tt.expected)
			}
		})
	}
}
