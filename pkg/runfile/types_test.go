package runfile

import (
	"context"
	"testing"

	"github.com/nxtcoder17/fastlog"
	"github.com/nxtcoder17/runfile/pkg/types"
)

func TestNewContext(t *testing.T) {
	baseCtx := &types.Context{
		Context: context.Background(),
		Logger:  fastlog.New(),
	}

	ctx := NewContext(baseCtx)

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	if ctx.Context != baseCtx {
		t.Error("NewContext did not properly embed the base context")
	}

	if ctx.taskTrail != nil {
		t.Error("NewContext should initialize taskTrail as nil")
	}

	// Test that we can access the embedded context methods
	if ctx.Logger == nil {
		t.Error("Cannot access Logger from embedded context")
	}
}