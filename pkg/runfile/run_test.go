package runfile

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

func Test_runTask(t *testing.T) {
	type args struct {
		rf   *Runfile
		args runTaskArgs
	}
	tests := []struct {
		name string
		args args
		want *Error
	}{
		{
			name: "1. Task Not Found",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{},
				},
				args: runTaskArgs{
					taskName: "sample",
				},
			},
			want: TaskNotFound,
		},

		{
			name: "1. Task Not Found",
			args: args{
				rf: &Runfile{
					Tasks: map[string]Task{},
				},
				args: runTaskArgs{
					taskName: "sample",
				},
			},
			want: TaskNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cf := context.WithTimeout(context.TODO(), 2*time.Second)
			defer cf()

			err := runTask(NewContext(ctx, slog.Default()), tt.args.rf, tt.args.args)
			if !reflect.DeepEqual(err, tt.want) {
				t.Errorf("runTask() = %v, want %v", err, tt.want)
			}
		})
	}
}
