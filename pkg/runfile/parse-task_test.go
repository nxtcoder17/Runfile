package runfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParsedRunfile_ParseTaskEnv(t *testing.T) {
	// Create a temporary directory for test dotenv files
	tmpDir := t.TempDir()

	// Create test .env files
	testEnvFile := filepath.Join(tmpDir, "test.env")
	if err := os.WriteFile(testEnvFile, []byte("TEST_VAR=test_value\nANOTHER_VAR=another_value"), 0o644); err != nil {
		t.Fatal(err)
	}

	prodEnvFile := filepath.Join(tmpDir, "prod.env")
	if err := os.WriteFile(prodEnvFile, []byte("PROD_VAR=production\nTEST_VAR=overridden"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		runfile   *ParsedRunfile
		taskName  string
		parentEnv map[string]string
		want      map[string]string
		wantErr   bool
	}{
		{
			name: "1. When task has no env, It should inherit parent and global env",
			runfile: &ParsedRunfile{
				Env: map[string]string{"GLOBAL": "global_value"},
				Tasks: map[string]Task{
					"simple": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
					},
				},
			},
			taskName:  "simple",
			parentEnv: map[string]string{"PARENT": "parent_value"},
			want: map[string]string{
				"PARENT": "parent_value",
				"GLOBAL": "global_value",
			},
			wantErr: false,
		},
		{
			name: "2. When task has env vars, It should merge with parent and global",
			runfile: &ParsedRunfile{
				Env: map[string]string{"GLOBAL": "global_value"},
				Tasks: map[string]Task{
					"with-env": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						Env: EnvExpr{
							"TASK_VAR": "task_value",
							"COMPUTED": "prefix_${GLOBAL}_suffix",
						},
					},
				},
			},
			taskName:  "with-env",
			parentEnv: map[string]string{"PARENT": "parent_value"},
			want: map[string]string{
				"PARENT":   "parent_value",
				"GLOBAL":   "global_value",
				"TASK_VAR": "task_value",
				"COMPUTED": "prefix_${GLOBAL}_suffix",
			},
			wantErr: false,
		},
		{
			name: "3. When task has dotenv file, It should load env from file",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"with-dotenv": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						DotEnv: []string{"test.env"},
					},
				},
			},
			taskName:  "with-dotenv",
			parentEnv: map[string]string{},
			want: map[string]string{
				"TEST_VAR":    "test_value",
				"ANOTHER_VAR": "another_value",
			},
			wantErr: false,
		},
		{
			name: "4. When task has multiple dotenv files, It should merge with later overriding earlier",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"multi-dotenv": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						DotEnv: []string{"test.env", "prod.env"},
					},
				},
			},
			taskName:  "multi-dotenv",
			parentEnv: map[string]string{},
			want: map[string]string{
				"TEST_VAR":    "overridden",
				"ANOTHER_VAR": "another_value",
				"PROD_VAR":    "production",
			},
			wantErr: false,
		},
		{
			name: "5. When task has absolute path dotenv, It should load the file",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"abs-dotenv": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						DotEnv: []string{testEnvFile},
					},
				},
			},
			taskName:  "abs-dotenv",
			parentEnv: map[string]string{},
			want: map[string]string{
				"TEST_VAR":    "test_value",
				"ANOTHER_VAR": "another_value",
			},
			wantErr: false,
		},
		{
			name: "6. When env var exists at multiple levels, It should follow precedence: parent < global < dotenv < task",
			runfile: &ParsedRunfile{
				Env: map[string]string{
					"VAR":         "global",
					"GLOBAL_ONLY": "global_only",
				},
				Tasks: map[string]Task{
					"precedence": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						DotEnv: []string{"test.env"},
						Env: EnvExpr{
							"VAR":       "task",
							"TASK_ONLY": "task_only",
						},
					},
				},
			},
			taskName: "precedence",
			parentEnv: map[string]string{
				"VAR":         "parent",
				"PARENT_ONLY": "parent_only",
			},
			want: map[string]string{
				"VAR":         "task",
				"PARENT_ONLY": "parent_only",
				"GLOBAL_ONLY": "global_only",
				"TASK_ONLY":   "task_only",
				"TEST_VAR":    "test_value",
				"ANOTHER_VAR": "another_value",
			},
			wantErr: false,
		},
		{
			name: "7. When task is not found, It should return error",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{},
			},
			taskName:  "nonexistent",
			parentEnv: map[string]string{},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "8. When dotenv file does not exist, It should return error",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"bad-dotenv": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						DotEnv: []string{"nonexistent.env"},
					},
				},
			},
			taskName:  "bad-dotenv",
			parentEnv: map[string]string{},
			want:      nil,
			wantErr:   true,
		},
	}

	ctx := NewTestContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.runfile.ParseTaskEnv(ctx, tt.taskName, tt.parentEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsedRunfile.ParseTaskEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsedRunfile.ParseTaskEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsedRunfile_ParseTaskEnv_WithRequires(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		runfile   *ParsedRunfile
		taskName  string
		parentEnv map[string]string
		wantErr   bool
	}{
		{
			name: "1. When requirement command succeeds, It should pass",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"with-requires": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						Requires: []*Requires{
							{
								Sh: stringPtr("true"),
							},
						},
					},
				},
			},
			taskName:  "with-requires",
			parentEnv: map[string]string{},
			wantErr:   false,
		},
		{
			name: "2. When requirement command fails, It should fail",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"failing-requires": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						Requires: []*Requires{
							{
								Sh: stringPtr("false"),
							},
						},
					},
				},
			},
			taskName:  "failing-requires",
			parentEnv: map[string]string{},
			wantErr:   true,
		},
		{
			name: "3. When requirements array contains nil, It should skip nil and continue",
			runfile: &ParsedRunfile{
				Env: map[string]string{},
				Tasks: map[string]Task{
					"nil-requires": {
						Metadata: struct {
							RunfilePath *string
							Namespace   string
						}{
							RunfilePath: stringPtr(filepath.Join(tmpDir, "Runfile")),
						},
						Requires: []*Requires{
							nil,
							{
								Sh: stringPtr("true"),
							},
						},
					},
				},
			},
			taskName:  "nil-requires",
			parentEnv: map[string]string{},
			wantErr:   false,
		},
	}

	ctx := NewTestContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.runfile.ParseTaskEnv(ctx, tt.taskName, tt.parentEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsedRunfile.ParseTaskEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

