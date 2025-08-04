package runfile

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/nxtcoder17/fastlog"
	fn "github.com/nxtcoder17/runfile/pkg/functions"
	"github.com/nxtcoder17/runfile/pkg/types"
)

func testParseCommandJsonEqual(t *testing.T, got, want *ParsedCommandJson) {
	if got == nil && want != nil || got != nil && want == nil {
		t.Errorf("parseCommand(),\n[.command] \n\tgot = %v\n\twant = %v", got, want)
		return
	}

	// t.Log("first", first, "err", err, "secondErr", secondErr, "condition", secondErr != (err != nil))

	if !reflect.DeepEqual(got.Command, want.Command) {
		t.Errorf("parseCommand(),\n[.command] \n\tgot = %v\n\twant = %v", fn.DefaultIfNil(got.Command, ""), fn.DefaultIfNil(want.Command, ""))
		return
	}

	if fmt.Sprint(got.Env) != fmt.Sprint(want.Env) {
		t.Errorf("parseCommand(),\n[.env] \n\tgot = %+v\n\twant = %+v", got.Env, want.Env)
		return
	}
}

func Test_parseCommand(t *testing.T) {
	type args struct {
		command any
		env     map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *ParsedCommandJson
		wantErr bool
	}{
		{
			name: "1. When command is simple string without env var, It should parse as command",
			args: args{
				command: "echo hello",
			},
			want: &ParsedCommandJson{
				Command: stringPtr("echo hello"),
			},
			wantErr: false,
		},
		{
			name: "2. When command is simple string with env var, It should include env",
			args: args{
				command: "echo hello",
				env:     map[string]string{"FOO": "bar"},
			},
			want: &ParsedCommandJson{
				Command: fn.Ptr("echo hello"),
				Env:     map[string]string{"FOO": "bar"},
			},
			wantErr: false,
		},
		{
			name: "3. When command uses json format with cmd key, It should merge environments",
			args: args{
				command: map[string]any{
					"cmd": "ls -la",
					"env": map[string]any{
						"KEY1": "value1",
					},
				},
				env: map[string]string{"FOO": "bar"},
			},
			want: &ParsedCommandJson{
				Command: stringPtr("ls -la"),
				Env: map[string]string{
					"FOO":  "bar",
					"KEY1": "value1",
				},
			},
			wantErr: false,
		},
		{
			name: "4. When command uses json format with run key, It should parse as task reference",
			args: args{
				command: map[string]any{
					"run": "build:prod",
					"env": map[string]any{
						"NODE_ENV": "production",
					},
				},
				env: map[string]string{"BASE": "value"},
			},
			want: &ParsedCommandJson{
				Run: stringPtr("build:prod"),
				Env: map[string]string{
					"BASE":     "value",
					"NODE_ENV": "production",
				},
			},
			wantErr: false,
		},
		{
			name: "5. When json format has no cmd or run key, It should fail",
			args: args{
				command: map[string]any{
					"invalid": "test",
				},
				env: map[string]string{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "6. When command type is invalid (number), It should fail",
			args: args{
				command: 123,
				env:     map[string]string{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "7. When both cmd and run keys exist, It should use run key",
			args: args{
				command: map[string]any{
					"cmd": "echo test",
					"run": "other:task",
				},
				env: map[string]string{},
			},
			want: &ParsedCommandJson{
				Run: stringPtr("other:task"),
				Env: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "8. When command is empty string, It should fail",
			args: args{
				command: "",
				env:     map[string]string{"TEST": "value"},
			},
			want: &ParsedCommandJson{
				Command: stringPtr(""),
				Env:     map[string]string{"TEST": "value"},
			},
			wantErr: true,
		},
		{
			name: "9. When json format has empty cmd string, It should fail",
			args: args{
				command: map[string]any{
					"cmd": "",
				},
				env: map[string]string{"TEST": "value"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "10. When json format has empty run target, It should fail",
			args: args{
				command: map[string]any{
					"run": "",
				},
				env: map[string]string{"TEST": "value"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "11. When environment is nil, It should handle gracefully",
			args: args{
				command: "test",
				env:     nil,
			},
			want: &ParsedCommandJson{
				Command: stringPtr("test"),
				Env:     nil,
			},
			wantErr: false,
		},
		{
			name: "12. When both cmd and run keys are empty strings, It should fail",
			args: args{
				command: map[string]any{
					"cmd": "",
					"run": "",
				},
				env: map[string]string{},
			},
			want:    nil,
			wantErr: true,
		},
	}

	ctx := NewContext(&types.Context{
		Context: context.TODO(),
		Logger:  fastlog.New(),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommand(ctx, tt.args.command, tt.args.env)
			if tt.wantErr && err == nil {
				t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if !tt.wantErr {
					t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			testParseCommandJsonEqual(t, got, tt.want)

			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("parseCommand() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

