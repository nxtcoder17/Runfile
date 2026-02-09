package resolver

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotEnvFilesInto(t *testing.T) {
	tests := []struct {
		name           string
		fileContents   []string // each entry becomes .env0, .env1, etc.
		initialStore   map[string]string
		expected       map[string]string
		wantErr        bool
		useRelPath     bool // if true, use relative path to trigger error
		useNonExistent bool // if true, use non-existent path
	}{
		{
			name:         "when files list is empty, it must pass",
			fileContents: []string{},
			initialStore: map[string]string{},
			expected:     map[string]string{},
			wantErr:      false,
		},
		{
			name:       "when file path is relative, it must fail",
			useRelPath: true,
			wantErr:    true,
		},
		{
			name:           "when file does not exist, it must fail",
			useNonExistent: true,
			wantErr:        true,
		},
		{
			name:         "when dotenv file has invalid syntax, it must return a parse error",
			fileContents: []string{"INVALID LINE WITHOUT EQUALS\nANOTHER BAD LINE"},
			initialStore: map[string]string{},
			expected:     map[string]string{},
			wantErr:      true,
		},
		{
			name: "when file is valid, it must parse all key-value pairs",
			fileContents: []string{
				`KEY1=value1
KEY2=value2
KEY3="quoted value"`,
			},
			initialStore: map[string]string{},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "quoted value",
			},
			wantErr: false,
		},
		{
			name: "when multiple files have same key, it must use value from last file",
			fileContents: []string{
				"KEY1=value1\nKEY2=original",
				"KEY2=overridden\nKEY3=value3",
			},
			initialStore: map[string]string{},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "overridden",
				"KEY3": "value3",
			},
			wantErr: false,
		},
		{
			name:         "when store has existing keys, it must preserve non-overlapping values",
			fileContents: []string{"NEW_KEY=new_value"},
			initialStore: map[string]string{"EXISTING": "existing_value"},
			expected: map[string]string{
				"EXISTING": "existing_value",
				"NEW_KEY":  "new_value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []string

			if tt.useRelPath {
				files = []string{"relative/path/.env"}
			} else if tt.useNonExistent {
				files = []string{"/non/existent/path/.env"}
			} else {
				tmpDir := t.TempDir()
				for i, content := range tt.fileContents {
					filename := fmt.Sprintf(".env%d", i)
					path := filepath.Join(tmpDir, filename)
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						t.Fatalf("failed to create temp file: %v", err)
					}
					files = append(files, path)
				}
			}

			store := make(map[string]string)
			for k, v := range tt.initialStore {
				store[k] = v
			}

			err := parseDotEnvFilesInto(store, files)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for k, v := range tt.expected {
				if store[k] != v {
					t.Errorf("expected %s=%q, got %s=%q", k, v, k, store[k])
				}
			}
		})
	}
}

func TestParseEnvInto(t *testing.T) {
	tests := []struct {
		name         string
		initialStore map[string]string
		envMap       map[string]any
		osEnvSetup   map[string]string // env vars to set before test
		expected     map[string]string
		wantErr      bool
	}{
		{
			name:         "when env map is empty, it must pass",
			initialStore: map[string]string{},
			envMap:       map[string]any{},
			expected:     map[string]string{},
			wantErr:      false,
		},
		{
			name:         "when value is string, it must store directly",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name:         "when value is numeric or boolean, it must convert to string",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"INT_KEY":   42,
				"FLOAT_KEY": 3.14,
				"BOOL_KEY":  true,
			},
			expected: map[string]string{
				"INT_KEY":   "42",
				"FLOAT_KEY": "3.14",
				"BOOL_KEY":  "true",
			},
			wantErr: false,
		},
		{
			name:         "when required is true and env var is missing, it must fail",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"REQUIRED_KEY": map[string]any{
					"required": true,
				},
			},
			wantErr: true,
		},
		{
			name:         "when required is false, it must pass",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"OPTIONAL_KEY": map[string]any{
					"required": false,
				},
			},
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:         "when required is not a boolean, it must fail",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"BAD_KEY": map[string]any{
					"required": "yes",
				},
			},
			wantErr: true,
		},
		{
			name:         "when sh key is present, it must evaluate and store trimmed output",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"SHELL_KEY": map[string]any{
					"sh": "echo hello",
				},
			},
			expected: map[string]string{
				"SHELL_KEY": "hello",
			},
			wantErr: false,
		},
		{
			name:         "when bash key is present, it must evaluate and store trimmed output",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"BASH_KEY": map[string]any{
					"bash": "echo -n world",
				},
			},
			expected: map[string]string{
				"BASH_KEY": "world",
			},
			wantErr: false,
		},
		{
			name:         "when shell script runs, it must have access to existing env vars",
			initialStore: map[string]string{"EXISTING": "from_store"},
			envMap: map[string]any{
				"DERIVED": map[string]any{
					"sh": "echo $EXISTING",
				},
			},
			expected: map[string]string{
				"EXISTING": "from_store",
				"DERIVED":  "from_store",
			},
			wantErr: false,
		},
		{
			name:         "when shell script value is not a string, it must fail",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"BAD_SHELL": map[string]any{
					"sh": 123,
				},
			},
			wantErr: true,
		},
		{
			name:         "when shell script exits with non-zero code, it must fail",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"FAIL_KEY": map[string]any{
					"sh": "exit 1",
				},
			},
			wantErr: true,
		},
		{
			name:         "when env var exists in store, it must skip evaluation",
			initialStore: map[string]string{"EXISTING": "original"},
			envMap: map[string]any{
				"EXISTING": map[string]any{
					"sh": "echo overwritten",
				},
			},
			expected: map[string]string{
				"EXISTING": "original",
			},
			wantErr: false,
		},
		{
			name:         "when env var exists in store, it must skip even for string values",
			initialStore: map[string]string{"EXISTING": "original"},
			envMap: map[string]any{
				"EXISTING": "overwritten",
			},
			expected: map[string]string{
				"EXISTING": "original",
			},
			wantErr: false,
		},
		{
			name:         "when env var exists in OS environment, it must skip evaluation",
			initialStore: map[string]string{},
			osEnvSetup:   map[string]string{"TEST_OS_ENV": "from_os"},
			envMap: map[string]any{
				"TEST_OS_ENV": map[string]any{
					"sh": "echo overwritten",
				},
			},
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:         "when required env var exists in OS environment, it must pass",
			initialStore: map[string]string{},
			osEnvSetup:   map[string]string{"REQUIRED_FROM_OS": "provided"},
			envMap: map[string]any{
				"REQUIRED_FROM_OS": map[string]any{
					"required": true,
				},
			},
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:         "when required env var exists in store, it must pass",
			initialStore: map[string]string{"REQUIRED_FROM_STORE": "provided"},
			envMap: map[string]any{
				"REQUIRED_FROM_STORE": map[string]any{
					"required": true,
				},
			},
			expected: map[string]string{
				"REQUIRED_FROM_STORE": "provided",
			},
			wantErr: false,
		},
		{
			name:         "when shell output has surrounding whitespace, it must trim it",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"TRIMMED": map[string]any{
					"sh": "echo '  spaces  '",
				},
			},
			expected: map[string]string{
				"TRIMMED": "spaces",
			},
			wantErr: false,
		},
		{
			name:         "when shell output has internal newlines, it must preserve them",
			initialStore: map[string]string{},
			envMap: map[string]any{
				"MULTILINE": map[string]any{
					"sh": "echo -e 'line1\\nline2'",
				},
			},
			expected: map[string]string{
				"MULTILINE": "line1\nline2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup OS env vars
			for k, v := range tt.osEnvSetup {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			store := make(map[string]string)
			maps.Copy(store, tt.initialStore)

			err := parseEnvInto(context.Background(), store, tt.envMap)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for k, v := range tt.expected {
				if store[k] != v {
					t.Errorf("expected %s=%q, got %s=%q", k, v, k, store[k])
				}
			}
		})
	}
}
