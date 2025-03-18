package parser

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/nxtcoder17/go.pkgs/log"
	fn "github.com/nxtcoder17/runfile/functions"
	"github.com/nxtcoder17/runfile/types"
)

func testParseCommandJsonEqual(t *testing.T, got, want *types.ParsedCommandJson) {
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
		prf     *types.ParsedRunfile
		taskEnv map[string]string
		command any
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ParsedCommandJson
		wantErr bool
	}{
		{
			name: "1. must pass with only command",
			args: args{
				prf:     &types.ParsedRunfile{},
				taskEnv: map[string]string{},
				command: "echo hi hello",
			},
			want: &types.ParsedCommandJson{
				Command: fn.New("echo hi hello"),
				Run:     nil,
				Env:     map[string]string{},
				If:      nil,
			},
			wantErr: false,
		},
		{
			name: "2. must fail with only run command with run target not found",
			args: args{
				prf:     &types.ParsedRunfile{},
				taskEnv: map[string]string{},
				command: map[string]any{
					"run": "build",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "3. must pass with run command, and target exists in runfile tasks",
			args: args{
				prf: &types.ParsedRunfile{
					Tasks: map[string]types.Task{
						"build": {
							Commands: []any{
								"echo from build",
							},
						},
					},
				},
				taskEnv: map[string]string{},
				command: map[string]any{
					"run": "build",
					"env": map[string]string{
						"k1": "v1",
					},
				},
			},
			want: &types.ParsedCommandJson{
				Command: nil,
				Run:     fn.New("build"),
				Env: map[string]string{
					"k1": "v1",
				},
				If: nil,
			},
			wantErr: false,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := types.Context{
				Context: context.TODO(),
				Logger:  log.New(),
			}

			got, err := parseCommand(ctx, tt.args.prf, tt.args.taskEnv, tt.args.command)
			if tt.wantErr != (err != nil) {
				t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			testParseCommandJsonEqual(t, got, tt.want)
		})
	}
}
