package runner

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nxtcoder17/fwatcher/pkg/executor"
	"github.com/nxtcoder17/runfile/types"
)

func TestCreateCommands(t *testing.T) {
	tests := []struct {
		prf types.ParsedRunfile
		pt  types.ParsedTask
		rta runTaskArgs

		want []executor.CommandGroup
	}{
		{
			prf: types.ParsedRunfile{},
			pt:  types.ParsedTask{},
		},
	}

	for _, tt := range tests {
		cg, err := createCommandGroups(Context{Context: context.TODO(), Logger: slog.Default()}, &tt.prf, &tt.pt, tt.rta)
		if err != nil {
			t.Error(err)
		}

		_ = cg
	}
}
