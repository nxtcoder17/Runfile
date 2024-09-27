package runfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
		if strings.Join(got.Shell, ",") != strings.Join(want.Shell, ",") {
			t.Logf("shell not equal")
			return false
		}

		slices.Sort(got.Environ)
		slices.Sort(want.Environ)

		if strings.Join(got.Environ, ",") != strings.Join(want.Environ, ",") {
			t.Logf("environ not equal")
			return false
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

	tests := []struct {
		name    string
		args    args
		want    *ParsedTask
		wantErr bool
	}{
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
				Environ: []string{
					"hello=hi",
					"k1=1",
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
				Environ: []string{
					"hello=hi",
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
				Environ: []string{
					"hello=world", // from dotenv
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
			name: "[unhappy/runfile] target task does not exist",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{},
				},
				taskName: "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(context.TODO(), tt.args.rf, tt.args.taskName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
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
