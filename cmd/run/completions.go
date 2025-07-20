package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/nxtcoder17/fastlog"
	"github.com/nxtcoder17/runfile/pkg/runfile"
	"github.com/nxtcoder17/runfile/pkg/types"
)

func generateShellCompletion(ctx context.Context, writer io.Writer, rfpath string) error {
	runfile, err := runfile.ParseFromFile(&types.Context{Context: ctx, Logger: fastlog.New()}, rfpath)
	if err != nil {
		slog.Error("parsing, got", "err", err)
		panic(err)
	}

	for k := range runfile.Tasks {
		fmt.Fprintf(writer, "%s\n", k)
	}

	return nil
}
