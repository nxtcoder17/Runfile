package parser

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nxtcoder17/runfile/types"
)

func pretty(v any) []byte {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return b
}

func Test_parseRunfile(t *testing.T) {
	type args struct {
		runfile *types.Runfile
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ParsedRunfile
		wantErr bool
	}{
		{
			name: "1. parsing env vars",
			args: args{
				runfile: &types.Runfile{
					Env: map[string]any{
						"env1": "value1",
					},
				},
			},
			want: &types.ParsedRunfile{
				Env: map[string]string{
					"env1": "value1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRunfile(tt.args.runfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRunfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRunfile:\ngot: %s\n\nwant:%s\n", pretty(got), pretty(tt.want))
				// t.Errorf("parseRunfile() = %v, want %v", got, tt.want)
			}
		})
	}
}
