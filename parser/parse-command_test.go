package parser

import (
	"reflect"
	"testing"

	"github.com/nxtcoder17/runfile/types"
)

func Test_parseCommand(t *testing.T) {
	type args struct {
		prf     *types.ParsedRunfile
		command any
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ParsedCommandJson
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommand(tt.args.prf, tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
