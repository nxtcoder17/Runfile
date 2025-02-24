package parser

import (
	"context"
	"reflect"
	"testing"

	"github.com/nxtcoder17/go.pkgs/log"
	. "github.com/nxtcoder17/runfile/types"
)

func Test_ParseEnvVars(t *testing.T) {
	type args struct {
		envVars    EnvVar
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
				envVars: EnvVar{
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
				envVars: EnvVar{
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
				envVars: EnvVar{
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
				envVars: EnvVar{
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
				envVars: EnvVar{
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
				envVars: EnvVar{
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
			got, err := parseEnvVars(Context{Context: context.TODO(), Logger: log.New(), TaskName: "test"}, tt.args.envVars, evaluationParams{
				Env: tt.args.testingEnv,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvVars():> got = %v, error = %v, wantErr %v", got, err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnvVars():> \n\tgot:\t%v,\n\twant:\t%v", got, tt.want)
			}
		})
	}
}
