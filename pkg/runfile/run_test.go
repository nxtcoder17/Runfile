package runfile

import (
	"reflect"
	"testing"
)

func Test_runTask(t *testing.T) {
	type args struct {
		ctx  Context
		rf   *Runfile
		args runTaskArgs
	}
	tests := []struct {
		name string
		args args
		want *Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runTask(tt.args.ctx, tt.args.rf, tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
