package runfile

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fn "github.com/nxtcoder17/runfile/pkg/functions"
)

func TestParseTask(t *testing.T) {
	type args struct {
		ctx      context.Context
		rf       *Runfile
		taskName string
	}

	areEqual := func(t *testing.T, got, want *ParsedTask) bool {
		if want == nil {
			return false
		}

		if strings.Join(got.Shell, ",") != strings.Join(want.Shell, ",") {
			t.Logf("shell not equal")
			return false
		}

		if got.Interactive != want.Interactive {
			t.Logf("interactive not equal")
			return false
		}

		if len(got.Env) != len(want.Env) {
			t.Logf("environments not equal")
			return false
		}

		gkeys := fn.MapKeys(got.Env)

		for _, k := range gkeys {
			v, ok := want.Env[k]
			if !ok || v != got.Env[k] {
				t.Logf("environments not equal")
				return false
			}
		}

		if got.WorkingDir != want.WorkingDir {
			t.Logf("working dir not equal")
			return false
		}

		if fmt.Sprintf("%#v", got.Commands) != fmt.Sprintf("%#v", want.Commands) {
			t.Logf("commands not equal:\n got:\t%#v\nwant:\t%#v", got.Commands, want.Commands)
			return false
		}

		return true
	}

	// for dotenv test
	dotenvTestFile, err := os.CreateTemp(os.TempDir(), ".env")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Fprintf(dotenvTestFile, "hello=world\n")
	dotenvTestFile.Close()

	type test struct {
		name    string
		args    args
		want    *ParsedTask
		wantErr bool
	}

	testRequires := []test{
		{
			name: "[requires] condition specified, but it neither has 'sh' or 'gotmpl' key",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           nil,
							ignoreSystemEnv: true,
							Requires: []*Requires{
								{},
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},

		{
			name: "[requires] condition specified, with gotmpl key",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           nil,
							ignoreSystemEnv: true,
							Requires: []*Requires{
								{
									GoTmpl: fn.New(`eq 5 5`),
								},
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},

		{
			name: "[requires] condition specified, with sh key",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           nil,
							ignoreSystemEnv: true,
							Requires: []*Requires{
								{
									Sh: fn.New(`echo hello`),
								},
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},

		{
			name: "[unhappy/requires] condition specified, with sh key",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           nil,
							ignoreSystemEnv: true,
							Requires: []*Requires{
								{
									Sh: fn.New(`echo hello && exit 1`),
								},
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []CommandJson{},
			},
			wantErr: true,
		},
	}

	testEnviroments := []test{
		{
			name: "[unhappy/env] required env, not provided",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: EnvVar{
								"hello": map[string]any{
									"required": true,
								},
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "[env] required env, provided",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: EnvVar{
								"hello": map[string]any{
									"required": true,
								},
							},
							DotEnv: []string{
								dotenvTestFile.Name(),
							},
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"hello": "world",
				},
				Commands: []CommandJson{},
			},
			wantErr: false,
		},
	}

	tests := []test{
		{
			name: "[shell] if not specified, defaults to [sh, -c]",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           nil,
							ignoreSystemEnv: true,
							Dir:             fn.New("."),
							Commands:        nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: ".",
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[shell] if specified, must be acknowledged",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           []string{"python", "-c"},
							ignoreSystemEnv: true,
							Dir:             fn.New("."),
							Commands:        nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"python", "-c"},
				WorkingDir: ".",
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[env] key: value",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							Shell:           []string{"sh", "-c"},
							ignoreSystemEnv: true,
							Env: map[string]any{
								"hello": "hi",
								"k1":    1,
							},
							Dir:      fn.New("."),
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell: []string{"sh", "-c"},
				Env: map[string]string{
					"hello": "hi",
					"k1":    "1",
				},
				WorkingDir: ".",
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[env] key: JSON object format",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: map[string]any{
								"hello": map[string]any{
									"sh": "echo hi",
								},
							},
							Dir:      fn.New("."),
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell: []string{"sh", "-c"},
				Env: map[string]string{
					"hello": "hi",
				},
				WorkingDir: ".",
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[unhappy/env] JSON object format [must throw err, when] sh key does not exist in value",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: map[string]any{
								"k1": map[string]any{
									"asdfasf": "asdfad",
								},
							},
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[unhappy/env] JSON object format [must throw err, when] sh (key)'s value is not a string",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: map[string]any{
								"k1": map[string]any{
									"sh": []string{"asdfsadf"},
								},
							},
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[dotenv] loads environment from given file",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							DotEnv: []string{
								dotenvTestFile.Name(),
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"hello": "world",
				},
				Commands: []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[unhappy/dotenv] throws err, when file does not exist",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							DotEnv: []string{
								"/tmp/env-aasfksadjfkl",
							},
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[unhappy/dotenv] throws err, when filepath exists [but] is not a file (might be a directory or something else)",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							DotEnv: []string{
								"/tmp",
							},
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[working_dir] if not specified, should be current working directory",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands:        nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[working_dir] when specified, must be acknowledged",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Dir:             fn.New("/tmp"),
							Commands:        nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: "/tmp",
				Commands:   []CommandJson{},
			},
			wantErr: false,
		},
		{
			name: "[unhappy/working_dir]  must throw err, when directory does not exist",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Dir:             fn.New("/tmp/xsdfjasdfkjdskfjasl"),
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[unhappy/working_dir] must throw err, when directory specified is not a directory (might be something else, or a file)",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Dir:             fn.New(filepath.Join(fn.Must(os.Getwd()), "task.go")),
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "[commands] string commands: single line",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hello",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands: []CommandJson{
					{Command: "echo hello"},
				},
			},
			wantErr: false,
		},

		{
			name: "[commands] string commands: multiline",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								`
echo "hello"
echo "hi"
`,
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands: []CommandJson{
					{
						Command: `
echo "hello"
echo "hi"
`,
					},
				},
			},
			wantErr: false,
		},

		{
			name: "[commands] JSON commands",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo i will call hello, now",
								map[string]any{
									"run": "hello",
								},
							},
						},
						"hello": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hello everyone",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands: []CommandJson{
					{Command: "echo i will call hello, now"},
					{Run: "hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "[unhappy/commands] JSON commands [must throw err, when] run target does not exist",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo i will call hello, now",
								map[string]any{
									"run": "hello",
								},
							},
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},

		{
			name: "[task] interactive task",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Interactive:     true,
							Commands: []any{
								"echo i will call hello, now",
								map[string]any{
									"run": "hello",
								},
							},
						},
						"hello": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hello everyone",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:       []string{"sh", "-c"},
				WorkingDir:  fn.Must(os.Getwd()),
				Interactive: true,
				Commands: []CommandJson{
					{Command: "echo i will call hello, now"},
					{Run: "hello"},
				},
			},
			wantErr: false,
		},
	}

	testGlobalEnvVars := []test{
		{
			name: "1. testing global env key-value item",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Env: map[string]any{
						"k1": "v1",
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"k1": "v1",
				},
				Commands: []CommandJson{
					{Command: "echo hi"},
				},
			},
			wantErr: false,
		},
		{
			name: "2. testing global env key-shell value",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Env: map[string]any{
						"k1": map[string]any{
							"sh": "echo hi",
						},
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"k1": "hi",
				},
				Commands: []CommandJson{
					{Command: "echo hi"},
				},
			},
			wantErr: false,
		},
		{
			name: "3. testing global env-var default value",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Env: map[string]any{
						"k1": map[string]any{
							"default": map[string]any{
								"value": "default-value",
							},
						},
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"k1": "default-value",
				},
				Commands: []CommandJson{
					{Command: "echo hi"},
				},
			},
			wantErr: false,
		},
		{
			name: "4. overriding global env var at task level",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Env: map[string]any{
						"k1": "v1",
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Env: EnvVar{
								"k1": "task-level-v1",
							},
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"k1": "task-level-v1",
				},
				Commands: []CommandJson{
					{Command: "echo hi"},
				},
			},
			wantErr: false,
		},

		{
			name: "5. required global env var",
			args: args{
				ctx: nil,
				rf: &Runfile{
					Env: map[string]any{
						"k1": map[string]any{
							"required": true,
						},
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want:    nil,
			wantErr: true,
		},
	}

	testGlobalDotEnv := []test{
		{
			name: "1. testing global env key-value item",
			args: args{
				ctx: nil,
				rf: &Runfile{
					DotEnv: []string{
						dotenvTestFile.Name(),
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Env: map[string]string{
					"hello": "world",
				},
				Commands: []CommandJson{
					{Command: "echo hi"},
				},
			},
			wantErr: false,
		},
		{
			name: "2. fails when dotenv file not found",
			args: args{
				ctx: nil,
				rf: &Runfile{
					DotEnv: []string{
						dotenvTestFile.Name() + "2",
					},
					Tasks: map[string]Task{
						"test": {
							ignoreSystemEnv: true,
							Commands: []any{
								"echo hi",
							},
						},
					},
				},
				taskName: "test",
			},
			want:    nil,
			wantErr: true,
		},
	}

	tests = append(tests, testRequires...)
	tests = append(tests, testEnviroments...)
	tests = append(tests, testGlobalEnvVars...)
	tests = append(tests, testGlobalDotEnv...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(NewContext(context.TODO(), slog.Default()), tt.args.rf, tt.args.rf.Tasks[tt.args.taskName])
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask(), got = %v, error = %v, wantErr %v", got, err, tt.wantErr)
				return
			}

			// if !reflect.DeepEqual(got, tt.want) {
			if !tt.wantErr {
				if !areEqual(t, got, tt.want) {
					t.Errorf("ParseTask():> \n\tgot:\t%v,\n\twant:\t%v", got, tt.want)
				}
			}
		})
	}
}
