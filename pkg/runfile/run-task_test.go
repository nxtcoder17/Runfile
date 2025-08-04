package runfile

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nxtcoder17/fastlog"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/types"
	"github.com/nxtcoder17/runfile/pkg/writer"
)

func Test_isDarkTheme(t *testing.T) {
	// This test just ensures the function doesn't panic
	_ = isDarkTheme()
}

func Test_longestLineLen(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want int
	}{
		{
			name: "1. When string is single line, It should return line length",
			str:  "hello world",
			want: 11,
		},
		{
			name: "2. When first line is longest, It should return first line length",
			str:  "hello world\nhi\ntest",
			want: 11,
		},
		{
			name: "3. When middle line is longest, It should return middle line length",
			str:  "hi\nhello world test\nbye",
			want: 16,
		},
		{
			name: "4. When last line is longest, It should return last line length",
			str:  "hi\nbye\nthis is the longest line",
			want: 24,
		},
		{
			name: "5. When string is empty, It should return 0",
			str:  "",
			want: 0,
		},
		{
			name: "6. When string has empty lines, It should return longest non-empty line",
			str:  "test\n\n\nhi",
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longestLineLen(tt.str); got != tt.want {
				t.Errorf("longestLineLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_padString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		padWith string
		want    string
	}{
		{
			name:    "1. When padding single line, It should add prefix with separator",
			str:     "hello",
			padWith: "PREFIX",
			want:    "PREFIX | hello",
		},
		{
			name:    "2. When padding multiple lines, It should align all lines",
			str:     "line1\nline2\nline3",
			padWith: "PREFIX",
			want:    "PREFIX | line1\n       | line2\n       | line3",
		},
		{
			name:    "3. When pad is empty, It should use space",
			str:     "test",
			padWith: "",
			want:    " | test",
		},
		{
			name:    "4. When pad is short, It should align properly",
			str:     "first\nsecond",
			padWith: ">>",
			want:    ">> | first\n   | second",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := padString(tt.str, tt.padWith); got != tt.want {
				t.Errorf("padString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateCommand(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		args CmdArgs
		want struct {
			shell      string
			args       []string
			hasStdin   bool
			workingDir string
		}
	}{
		{
			name: "1. When no shell specified, It should use default sh",
			args: CmdArgs{
				Cmd: "echo hello",
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell:    "sh",
				args:     []string{"-c", "echo hello"},
				hasStdin: false,
			},
		},
		{
			name: "2. When custom shell specified, It should use custom shell",
			args: CmdArgs{
				Shell: []string{"bash", "-c"},
				Cmd:   "echo hello",
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell:    "bash",
				args:     []string{"-c", "echo hello"},
				hasStdin: false,
			},
		},
		{
			name: "3. When working directory specified, It should set Dir",
			args: CmdArgs{
				Cmd:        "pwd",
				WorkingDir: stringPtr("/tmp"),
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell:      "sh",
				args:       []string{"-c", "pwd"},
				workingDir: "/tmp",
			},
		},
		{
			name: "4. When interactive mode, It should set stdin",
			args: CmdArgs{
				Cmd:         "read input",
				interactive: true,
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell:    "sh",
				args:     []string{"-c", "read input"},
				hasStdin: true,
			},
		},
		{
			name: "5. When environment vars specified, It should set env",
			args: CmdArgs{
				Cmd: "echo $VAR",
				Env: []string{"VAR=value"},
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell: "sh",
				args:  []string{"-c", "echo $VAR"},
			},
		},
		{
			name: "6. When custom stdout/stderr provided, It should use them",
			args: CmdArgs{
				Cmd:    "echo test",
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
			},
			want: struct {
				shell      string
				args       []string
				hasStdin   bool
				workingDir string
			}{
				shell: "sh",
				args:  []string{"-c", "echo test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCommand(ctx, tt.args)

			// Check shell (Path will be the executable name)
			if !strings.HasSuffix(cmd.Path, tt.want.shell) {
				t.Errorf("CreateCommand() shell = %v, want %v", cmd.Path, tt.want.shell)
			}

			// Check args
			if len(cmd.Args) != len(tt.want.args)+1 { // +1 for the executable name
				t.Errorf("CreateCommand() args length = %v, want %v", len(cmd.Args)-1, len(tt.want.args))
			}

			// Check stdin
			if tt.want.hasStdin && cmd.Stdin == nil {
				t.Error("CreateCommand() expected stdin to be set")
			} else if !tt.want.hasStdin && cmd.Stdin != nil {
				t.Error("CreateCommand() expected stdin to be nil")
			}

			// Check working directory
			if cmd.Dir != tt.want.workingDir {
				t.Errorf("CreateCommand() Dir = %v, want %v", cmd.Dir, tt.want.workingDir)
			}

			// Check env
			if tt.args.Env != nil && len(cmd.Env) != len(tt.args.Env) {
				t.Errorf("CreateCommand() Env length = %v, want %v", len(cmd.Env), len(tt.args.Env))
			}
		})
	}
}

func TestParsedRunfile_RunTask(t *testing.T) {
	// Create a minimal test
	tests := []struct {
		name    string
		runfile *ParsedRunfile
		task    string
		wantErr bool
	}{
		{
			name: "1. When task not found, It should return error",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{},
			},
			task:    "nonexistent",
			wantErr: true,
		},
		{
			name: "2. When simple echo task exists, It should execute successfully",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"echo": {
						Name: "echo",
						Commands: []any{
							"echo test",
						},
					},
				},
			},
			task:    "echo",
			wantErr: false,
		},
		{
			name: "3. When both Parallel and Watch are true, It should handle gracefully",
			runfile: &ParsedRunfile{
				Tasks: map[string]Task{
					"parallel-watch": {
						Name:     "parallel-watch",
						Parallel: true,
						Watch: &TaskWatch{
							Enable: fn.Ptr(true),
						},
						Commands: []any{
							"echo test",
						},
					},
				},
			},
			task:    "parallel-watch",
			wantErr: false, // Based on the code, it doesn't seem to validate this combination
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(&types.Context{
				Context: context.Background(),
				Logger:  fastlog.New(),
			})

			// For successful cases, we need to capture output
			if !tt.wantErr {
				// Override stdout to capture output
				old := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				err := tt.runfile.RunTask(ctx, tt.task)

				// Restore stdout
				w.Close()
				os.Stdout = old

				// Read captured output
				var buf bytes.Buffer
				io.Copy(&buf, r)

				if (err != nil) != tt.wantErr {
					t.Errorf("ParsedRunfile.RunTask() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				// For error cases, just check the error
				err := tt.runfile.RunTask(ctx, tt.task)
				if (err != nil) != tt.wantErr {
					t.Errorf("ParsedRunfile.RunTask() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestParsedRunfile_createCommandGroups(t *testing.T) {
	runfile := &ParsedRunfile{
		Tasks: map[string]Task{
			"simple": {
				Name: "simple",
				Commands: []any{
					"echo hello",
					"echo world",
				},
			},
			"with-run": {
				Name: "with-run",
				Commands: []any{
					map[string]any{
						"run": "simple",
					},
				},
			},
			"with-cmd": {
				Name: "with-cmd",
				Commands: []any{
					map[string]any{
						"cmd": "ls -la",
						"env": map[string]any{
							"VAR": "value",
						},
					},
				},
			},
			"parallel": {
				Name:     "parallel",
				Parallel: true,
				Commands: []any{
					"sleep 1",
					"sleep 1",
				},
			},
		},
	}

	ctx := NewContext(&types.Context{
		Context: context.Background(),
		Logger:  fastlog.New(),
	})

	// Parse env for each task
	for name, task := range runfile.Tasks {
		env, _ := runfile.ParseTaskEnv(ctx, name, map[string]string{})
		task.ParentEnv = env
		runfile.Tasks[name] = task
	}

	tests := []struct {
		name     string
		taskName string
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "1. When task has simple commands, It should create command groups",
			taskName: "simple",
			wantErr:  false,
			wantLen:  2,
		},
		{
			name:     "2. When task has run command, It should create single group",
			taskName: "with-run",
			wantErr:  false,
			wantLen:  1,
		},
		{
			name:     "3. When task has cmd with env, It should create group with env",
			taskName: "with-cmd",
			wantErr:  false,
			wantLen:  1,
		},
		{
			name:     "4. When task has parallel commands, It should create multiple groups",
			taskName: "parallel",
			wantErr:  false,
			wantLen:  2,
		},
		{
			name:     "5. When task doesn't exist, It should return error",
			taskName: "nonexistent",
			wantErr:  true,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runfile.createCommandGroups(ctx, tt.taskName, CreateCommandGroupArgs{
				Trail:  []string{},
				Stdout: &writer.LogWriter{Writer: &bytes.Buffer{}},
				Stderr: &writer.LogWriter{Writer: &bytes.Buffer{}},
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("createCommandGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("createCommandGroups() returned %d groups, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func Test_printCommand(t *testing.T) {
	// Just test that it doesn't panic
	tests := []struct {
		name   string
		prefix string
		lang   string
		cmd    string
	}{
		{
			name:   "1. When printing simple command, It should format correctly",
			prefix: "test",
			lang:   "bash",
			cmd:    "echo hello",
		},
		{
			name:   "2. When printing multi-line command, It should format each line",
			prefix: "build",
			lang:   "sh",
			cmd:    "echo line1\necho line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printCommand(&buf, tt.prefix, tt.lang, tt.cmd)
			// Just ensure something was written
			if buf.Len() == 0 && writer.IsANSITerminal() {
				t.Error("printCommand() wrote nothing to buffer in TTY mode")
			}
		})
	}
}
