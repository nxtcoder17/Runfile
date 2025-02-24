package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/nxtcoder17/go.pkgs/log"
	"github.com/nxtcoder17/runfile/parser"
	"github.com/nxtcoder17/runfile/types"
)

func generateShellCompletion(ctx context.Context, writer io.Writer, rfpath string) error {
	runfile, err := parser.ParseRunfile(types.NewContext(ctx, log.New()), rfpath)
	if err != nil {
		slog.Error("parsing, got", "err", err)
		panic(err)
	}

	for k := range runfile.Tasks {
		fmt.Fprintf(writer, "%s\n", k)
	}

	return nil
}
