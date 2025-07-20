package runfile

import (
	"context"
	"reflect"
	"testing"

	"github.com/nxtcoder17/fastlog"
	"github.com/nxtcoder17/runfile/pkg/types"
)

func Test_ParseEnvExprs(t *testing.T) {
	type args struct {
		envVars    types.EnvExpr
		testingEnv map[string]string
	}

	type test struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}

	tests := []test{
		{
			name: "1. must fail [when] required env is not provided",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"required": true,
					},
				},
				testingEnv: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "2. must pass [when] required env is provided",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"required": true,
					},
				},
				testingEnv: map[string]string{
					"hello": "world",
				},
			},
			want: map[string]string{
				"hello": "world",
			},
			wantErr: false,
		},
		{
			name: "3. must fail [when] default not provided",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"required": true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "4. must pass [when] default value is provided",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"default": "world",
					},
				},
				testingEnv: nil,
			},
			want: map[string]string{
				"hello": "world",
			},
			wantErr: false,
		},
		{
			name: "5. must fail [when] default sh command exits with non-zero",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"default": map[string]any{
							"sh": "exit 1",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "6. must pass [when] default sh command exits with zero",
			args: args{
				envVars: types.EnvExpr{
					"hello": map[string]any{
						"default": map[string]any{
							"sh": "echo hi",
						},
					},
				},
			},
			want: map[string]string{
				"hello": "hi",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvVars(types.Context{Context: context.TODO(), Logger: fastlog.New()}, tt.args.envVars, tt.args.testingEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parsetypes.EnvExprs():> got = %v, error = %v, wantErr %v", got, err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsetypes.EnvExprs():> \n\tgot:\t%v,\n\twant:\t%v", got, tt.want)
			}
		})
	}
}
