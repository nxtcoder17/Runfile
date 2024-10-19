package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/nxtcoder17/runfile/pkg/runfile"
)

func generateShellCompletion(_ context.Context, writer io.Writer, rfpath string) error {
	// if c.NArg() > 0 {
	// 	return nil
	// }

	// runfilePath, err := locateRunfile(c)
	// if err != nil {
	// 	slog.Error("locating runfile", "err", err)
	// 	panic(err)
	// }

	runfile, err := runfile.Parse(rfpath)
	if err != nil {
		slog.Error("parsing, got", "err", err)
		panic(err)
	}

	for k := range runfile.Tasks {
		fmt.Fprintf(writer, "%s\n", k)
	}

	m, err := runfile.ParseIncludes()
	if err != nil {
		slog.Error("parsing, got", "err", err)
		panic(err)
	}

	for k, v := range m {
		for tn := range v.Runfile.Tasks {
			fmt.Fprintf(writer, "%s:%s\n", k, tn)
		}
	}

	return nil
}
