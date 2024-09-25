package runfile

import (
	"context"
	"testing"
)

func TestRunFile_Run(t *testing.T) {
	type fields struct {
		attrs   attrs
		Version string
		Tasks   map[string]TaskSpec
	}
	type args struct {
		ctx   context.Context
		tasks []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runfile{
				attrs:   tt.fields.attrs,
				Version: tt.fields.Version,
				Tasks:   tt.fields.Tasks,
			}
			if err := r.Run(tt.args.ctx, tt.args.tasks); (err != nil) != tt.wantErr {
				t.Errorf("RunFile.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
