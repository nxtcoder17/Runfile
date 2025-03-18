package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
	. "github.com/nxtcoder17/runfile/types"
)

func Test_ParseTask(t *testing.T) {
	type args struct {
		ctx      context.Context
		rf       *ParsedRunfile
		taskName string
	}

	compareParsedTasks := func(t *testing.T, got, want *ParsedTask) {
		if got == nil && want != nil || got != nil && want == nil {
			t.Errorf("ParseTask(), \n\tgot = %v\n\twant = %v", got, want)
			return
		}

		if strings.Join(got.Shell, ",") != strings.Join(want.Shell, ",") {
			t.Errorf("ParseTask(),\n[shell not equal]\n\tgot = %+v\n\twant = %+v", got.Shell, want.Shell)
			return
		}

		if got.Interactive != want.Interactive {
			t.Errorf("ParseTask(),\n[.interactive not equal]\n\tgot = %+v\n\twant = %+v", got.Interactive, want.Interactive)
			return
		}

		if fmt.Sprint(got.Env) != fmt.Sprint(want.Env) {
			t.Errorf("ParseTask(),\n[.Env not equal]\n\tgot = %+v\n\twant = %+v", got.Env, want.Env)
			return
		}

		if got.WorkingDir != want.WorkingDir {
			t.Errorf("ParseTask(),\n[.WorkingDir not equal]\n\tgot = %+v\n\twant = %+v", got.WorkingDir, want.WorkingDir)
			return
		}

		if len(got.Commands) != len(want.Commands) {
			t.Errorf("ParseTask(),\n[len(.Commands) not equal]\n\tgot = %+v\n\twant = %+v", len(got.Commands), len(want.WorkingDir))
			return
		}

		for i := 0; i < len(got.Commands); i++ {
			testParseCommandJsonEqual(t, &got.Commands[i], &want.Commands[i])
			return
		}
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

	// testRequires := []test{
	// 	{
	// 		name: "1: [requires] condition specified, but it neither has 'sh' or 'gotmpl' key",
	// 		args: args{
	// 			rf: &ParsedRunfile{
	// 				Tasks: map[string]Task{
	// 					"test": {
	// 						Shell: nil,
	// 						// ignoreSystemEnv: true,
	// 						Requires: []*Requires{
	// 							{},
	// 						},
	// 						Commands: nil,
	// 					},
	// 				},
	// 			},
	// 			taskName: "test",
	// 		},
	// 		want: &ParsedTask{
	// 			Shell:      []string{"sh", "-c"},
	// 			WorkingDir: fn.Must(os.Getwd()),
	// 			Commands:   []ParsedCommandJson{},
	// 		},
	// 		wantErr: false,
	// 	},
	//
	// 	{
	// 		name: "[requires] condition specified, with gotmpl key",
	// 		args: args{
	// 			rf: &ParsedRunfile{
	// 				Tasks: map[string]Task{
	// 					"test": {
	// 						Shell: nil,
	// 						Requires: []*Requires{
	// 							{
	// 								GoTmpl: fn.New(`eq 5 5`),
	// 							},
	// 						},
	// 						Commands: nil,
	// 					},
	// 				},
	// 			},
	// 			taskName: "test",
	// 		},
	// 		want: &ParsedTask{
	// 			Shell:      []string{"sh", "-c"},
	// 			WorkingDir: fn.Must(os.Getwd()),
	// 			Commands:   []ParsedCommandJson{},
	// 		},
	// 		wantErr: false,
	// 	},
	//
	// 	{
	// 		name: "[requires] condition specified, with sh key",
	// 		args: args{
	// 			rf: &ParsedRunfile{
	// 				Tasks: map[string]Task{
	// 					"test": {
	// 						Shell: nil,
	// 						Requires: []*Requires{
	// 							{
	// 								Sh: fn.New(`echo hello`),
	// 							},
	// 						},
	// 						Commands: nil,
	// 					},
	// 				},
	// 			},
	// 			taskName: "test",
	// 		},
	// 		want: &ParsedTask{
	// 			Shell:      []string{"sh", "-c"},
	// 			WorkingDir: fn.Must(os.Getwd()),
	// 			Commands:   []ParsedCommandJson{},
	// 		},
	// 		wantErr: false,
	// 	},
	//
	// 	{
	// 		name: "[unhappy/requires] condition specified, with sh key",
	// 		args: args{
	// 			rf: &ParsedRunfile{
	// 				Tasks: map[string]Task{
	// 					"test": {
	// 						Shell: nil,
	// 						Requires: []*Requires{
	// 							{
	// 								Sh: fn.New(`echo hello && exit 1`),
	// 							},
	// 						},
	// 						Commands: nil,
	// 					},
	// 				},
	// 			},
	// 			taskName: "test",
	// 		},
	// 		want: &ParsedTask{
	// 			Shell:      []string{"sh", "-c"},
	// 			WorkingDir: fn.Must(os.Getwd()),
	// 			Commands:   []ParsedCommandJson{},
	// 		},
	// 		wantErr: true,
	// 	},
	// }

	tests := []test{
		{
			name: "1. [shell] if not specified, defaults to [sh, -c]",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Shell:    nil,
							Dir:      fn.New("."),
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: ".",
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "2. [shell] if specified, must be acknowledged",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Shell:    []string{"python", "-c"},
							Dir:      fn.New("."),
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"python", "-c"},
				WorkingDir: ".",
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "3. [env] key: value",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Shell: []string{"sh", "-c"},
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
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "4. [env] key: JSON object format",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Env: EnvVar{
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
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "5. [unhappy/env] JSON object format [must throw err, when] sh key does not exist in value",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
			name: "6. [unhappy/env] JSON object format [must throw err, when] sh (key)'s value is not a string",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
			name: "7. [dotenv] loads environment from given file",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
				Commands: []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "8. [unhappy/dotenv] throws err, when file does not exist",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
			name: "9. [unhappy/dotenv] throws err, when filepath exists [but] is not a file (might be a directory or something else)",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
			name: "10. [working_dir] if not specified, should be current working directory",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: fn.Must(os.Getwd()),
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "11. [working_dir] when specified, must be acknowledged",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Dir:      fn.New("/tmp"),
							Commands: nil,
						},
					},
				},
				taskName: "test",
			},
			want: &ParsedTask{
				Shell:      []string{"sh", "-c"},
				WorkingDir: "/tmp",
				Commands:   []ParsedCommandJson{},
			},
			wantErr: false,
		},
		{
			name: "12. [unhappy/working_dir]  must throw err, when directory does not exist",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Dir: fn.New("/tmp/xsdfjasdfkjdskfjasl"),
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "13. [unhappy/working_dir] must throw err, when directory specified is not a directory (might be something else, or a file)",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Dir: fn.New(filepath.Join(fn.Must(os.Getwd()), "task.go")),
						},
					},
				},
				taskName: "test",
			},
			wantErr: true,
		},
		{
			name: "14. [commands] string commands: single line",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
				Commands: []ParsedCommandJson{
					{
						Command: fn.New("echo hello"),
					},
				},
			},
			wantErr: false,
		},

		{
			name: "15. [commands] string commands: multiline",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
				Commands: []ParsedCommandJson{
					{
						Command: fn.New(`
echo "hello"
echo "hi"
`),
					},
				},
			},
			wantErr: false,
		},

		{
			name: "16. [commands] JSON commands",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Commands: []any{
								"echo i will call hello, now",
								map[string]any{
									"run": "hello",
								},
							},
						},
						"hello": {
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
				Commands: []ParsedCommandJson{
					{
						Command: fn.New("echo i will call hello, now"),
					},
					{
						Run: fn.New("hello"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "17. [unhappy/commands] JSON commands [must throw err, when] run target does not exist",
			args: args{
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
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
			name: "18. [task] interactive task",
			args: args{
				ctx: nil,
				rf: &ParsedRunfile{
					Tasks: map[string]Task{
						"test": {
							Interactive: true,
							Commands: []any{
								"echo i will call hello, now",
								map[string]any{
									"run": "hello",
								},
							},
						},
						"hello": {
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
				Commands: []ParsedCommandJson{
					{Command: fn.New("echo i will call hello, now")},
					{Run: fn.New("hello")},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(types.Context{Context: t.Context()}, tt.args.rf, tt.args.rf.Tasks[tt.args.taskName])
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask(), got = %v, error = %v, wantErr %v", got, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			compareParsedTasks(t, got, tt.want)
		})
	}
}
