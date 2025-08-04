package runfile

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nxtcoder17/fastlog"
	"github.com/nxtcoder17/runfile/pkg/types"
)

func TestParseFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Define lib runfile path first
	libRunfilePath := filepath.Join(tmpDir, "lib.Runfile")

	// Create test runfile
	runfileContent := `
version: 1.0.0
env:
  GLOBAL_VAR: global_value
  COMPUTED: "prefix_${USER}_suffix"

dotEnv:
  - test.env

tasks:
  build:
    shell: bash
    env:
      BUILD_MODE: production
    cmd:
      - echo "Building..."
      - echo "Done"
  
  test:
    shell: sh
    cmd:
      - go test ./...

includes:
  lib:
    runfile: ` + libRunfilePath + `
`

	runfilePath := filepath.Join(tmpDir, "Runfile")
	if err := os.WriteFile(runfilePath, []byte(runfileContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test.env file
	testEnvContent := "ENV_VAR=env_value\nANOTHER=test"
	testEnvPath := filepath.Join(tmpDir, "test.env")
	if err := os.WriteFile(testEnvPath, []byte(testEnvContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create lib.Runfile for includes test
	libRunfileContent := `
tasks:
  compile:
    cmd:
      - echo "Compiling library..."
  clean:
    dir: build
    cmd:
      - rm -rf *
`
	if err := os.WriteFile(libRunfilePath, []byte(libRunfileContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &types.Context{
		Context: context.TODO(),
		Logger:  fastlog.New(),
	}

	tests := []struct {
		name    string
		file    string
		wantErr bool
		check   func(t *testing.T, pr *ParsedRunfile)
	}{
		{
			name:    "1. When runfile is valid, It should parse successfully",
			file:    runfilePath,
			wantErr: false,
			check: func(t *testing.T, pr *ParsedRunfile) {
				// Check tasks exist
				if _, ok := pr.Tasks["build"]; !ok {
					t.Error("Expected 'build' task to exist")
				}
				if _, ok := pr.Tasks["test"]; !ok {
					t.Error("Expected 'test' task to exist")
				}
				if _, ok := pr.Tasks["lib:compile"]; !ok {
					t.Error("Expected 'lib:compile' task to exist from includes")
				}
				if _, ok := pr.Tasks["lib:clean"]; !ok {
					t.Error("Expected 'lib:clean' task to exist from includes")
				}

				// Check env vars
				if pr.Env["GLOBAL_VAR"] != "global_value" {
					t.Errorf("Expected GLOBAL_VAR=global_value, got %s", pr.Env["GLOBAL_VAR"])
				}
				if pr.Env["ENV_VAR"] != "env_value" {
					t.Errorf("Expected ENV_VAR=env_value from dotenv, got %s", pr.Env["ENV_VAR"])
				}

				// Check task properties
				buildTask := pr.Tasks["build"]
				if buildTask.Name != "build" {
					t.Errorf("Expected task name 'build', got %s", buildTask.Name)
				}
				if len(buildTask.Commands) != 2 {
					t.Errorf("Expected 2 commands in build task, got %d", len(buildTask.Commands))
				}

				// Check included task properties
				libCleanTask := pr.Tasks["lib:clean"]
				if libCleanTask.Name != "lib:clean" {
					t.Errorf("Expected task name 'lib:clean', got %s", libCleanTask.Name)
				}
				if libCleanTask.Metadata.Namespace != "lib" {
					t.Errorf("Expected namespace 'lib', got %s", libCleanTask.Metadata.Namespace)
				}
				if libCleanTask.Dir == nil || *libCleanTask.Dir != "build" {
					t.Error("Expected dir to be 'build'")
				}
			},
		},
		{
			name:    "2. When file doesn't exist, It should return error",
			file:    filepath.Join(tmpDir, "nonexistent.yaml"),
			wantErr: true,
		},
		{
			name: "3. When yaml is invalid, It should return error",
			file: func() string {
				invalidPath := filepath.Join(tmpDir, "invalid.yaml")
				os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644)
				return invalidPath
			}(),
			wantErr: true,
		},
		{
			name: "4. When include file doesn't exist, It should return error",
			file: func() string {
				invalidIncludePath := filepath.Join(tmpDir, "invalid-include.yaml")
				content := `
tasks:
  test:
    cmd: echo test
includes:
  bad:
    runfile: nonexistent.yaml
`
				os.WriteFile(invalidIncludePath, []byte(content), 0644)
				return invalidIncludePath
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFromFile(ctx, tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestRunfile_resolveEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test .env files
	env1Content := "VAR1=value1\nVAR2=value2"
	env1Path := filepath.Join(tmpDir, "env1.env")
	if err := os.WriteFile(env1Path, []byte(env1Content), 0644); err != nil {
		t.Fatal(err)
	}

	env2Content := "VAR2=overridden\nVAR3=value3"
	env2Path := filepath.Join(tmpDir, "env2.env")
	if err := os.WriteFile(env2Path, []byte(env2Content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &types.Context{
		Context: context.TODO(),
		Logger:  fastlog.New(),
	}

	tests := []struct {
		name    string
		runfile *Runfile
		want    map[string]string
		wantErr bool
	}{
		{
			name: "1. When no env or dotenv specified, It should return empty map",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
			},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "2. When only env vars specified, It should parse env vars",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				Env: types.EnvExpr{
					"KEY1": "value1",
					"KEY2": "value2",
				},
			},
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
		{
			name: "3. When only dotenv files specified, It should load env from files",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				DotEnv:   []string{env1Path},
			},
			want: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			wantErr: false,
		},
		{
			name: "4. When dotenv has relative paths, It should resolve from runfile dir",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				DotEnv:   []string{"env1.env", "env2.env"},
			},
			want: map[string]string{
				"VAR1": "value1",
				"VAR2": "overridden",
				"VAR3": "value3",
			},
			wantErr: false,
		},
		{
			name: "5. When both env and dotenv specified, It should have env override dotenv",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				DotEnv:   []string{env1Path},
				Env: types.EnvExpr{
					"VAR1": "env_override",
					"NEW":  "new_value",
				},
			},
			want: map[string]string{
				"VAR1": "env_override",
				"VAR2": "value2",
				"NEW":  "new_value",
			},
			wantErr: false,
		},
		{
			name: "6. When dotenv file doesn't exist, It should return error",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				DotEnv:   []string{"nonexistent.env"},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.runfile.resolveEnv(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runfile.resolveEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Runfile.resolveEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunfile_resolveIncludedTasks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included runfiles
	lib1Content := `
env:
  LIB1_VAR: lib1_value

tasks:
  build:
    dir: src
    cmd:
      - echo "Building lib1"
  test:
    cmd:
      - echo "Testing lib1"
`
	lib1Path := filepath.Join(tmpDir, "lib1.yaml")
	if err := os.WriteFile(lib1Path, []byte(lib1Content), 0644); err != nil {
		t.Fatal(err)
	}

	lib2Content := `
tasks:
  deploy:
    cmd:
      - echo "Deploying lib2"
`
	lib2Path := filepath.Join(tmpDir, "lib2.yaml")
	if err := os.WriteFile(lib2Path, []byte(lib2Content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &types.Context{
		Context: context.TODO(),
		Logger:  fastlog.New(),
	}

	tests := []struct {
		name    string
		runfile *Runfile
		wantErr bool
		check   func(t *testing.T, tasks map[string]Task)
	}{
		{
			name: "1. When no includes specified, It should return empty tasks",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
			},
			wantErr: false,
			check: func(t *testing.T, tasks map[string]Task) {
				if len(tasks) != 0 {
					t.Errorf("Expected no tasks, got %d", len(tasks))
				}
			},
		},
		{
			name: "2. When single include specified, It should namespace tasks",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				Includes: map[string]IncludeSpec{
					"lib1": {
						Runfile: lib1Path,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, tasks map[string]Task) {
				if len(tasks) != 2 {
					t.Errorf("Expected 2 tasks, got %d", len(tasks))
				}

				buildTask, ok := tasks["lib1:build"]
				if !ok {
					t.Error("Expected lib1:build task")
				} else {
					if buildTask.Name != "lib1:build" {
						t.Errorf("Expected task name lib1:build, got %s", buildTask.Name)
					}
					if buildTask.Metadata.Namespace != "lib1" {
						t.Errorf("Expected namespace lib1, got %s", buildTask.Metadata.Namespace)
					}
					if buildTask.Dir == nil || *buildTask.Dir != "src" {
						t.Error("Expected dir to be 'src'")
					}
				}

				if _, ok := tasks["lib1:test"]; !ok {
					t.Error("Expected lib1:test task")
				}
			},
		},
		{
			name: "3. When multiple includes specified, It should namespace all tasks",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				Includes: map[string]IncludeSpec{
					"lib1": {
						Runfile: lib1Path,
					},
					"lib2": {
						Runfile: lib2Path,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, tasks map[string]Task) {
				expectedTasks := []string{"lib1:build", "lib1:test", "lib2:deploy"}
				if len(tasks) != len(expectedTasks) {
					t.Errorf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
				}

				for _, taskName := range expectedTasks {
					if _, ok := tasks[taskName]; !ok {
						t.Errorf("Expected task %s to exist", taskName)
					}
				}
			},
		},
		{
			name: "4. When include has dir override, It should adjust task dirs",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				Includes: map[string]IncludeSpec{
					"lib": {
						Runfile: lib1Path,
						Dir:     "libs/lib1",
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, tasks map[string]Task) {
				buildTask, ok := tasks["lib:build"]
				if !ok {
					t.Error("Expected lib:build task")
				} else {
					if buildTask.Dir == nil {
						t.Error("Expected dir to be set")
					} else if *buildTask.Dir != filepath.Join("libs/lib1", "src") {
						t.Errorf("Expected dir to be 'libs/lib1/src', got %s", *buildTask.Dir)
					}
				}
			},
		},
		{
			name: "5. When include file doesn't exist, It should return error",
			runfile: &Runfile{
				Filepath: filepath.Join(tmpDir, "Runfile"),
				Includes: map[string]IncludeSpec{
					"bad": {
						Runfile: filepath.Join(tmpDir, "nonexistent.yaml"),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.runfile.resolveIncludedTasks(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runfile.resolveIncludedTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}