package runner

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/nxtcoder17/runfile/types"
)

type PrefixedWriter struct {
	w      io.Writer
	prefix []byte
	buf    *bytes.Buffer
	render func([]byte) []byte
}

func (pw *PrefixedWriter) Write(p []byte) (int, error) {
	defer pw.buf.Reset()
	n, err := pw.buf.Write(p)
	if err != nil {
		return n, err
	}

	for {
		line, err := pw.buf.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			pw.buf.Reset()
			pw.buf.Write(pw.render(line))
			break
		}

		if _, err := pw.w.Write(append(pw.prefix, pw.render(line)...)); err != nil {
			return n, err
		}
	}
	return n, nil
}

var _ io.Writer = (*PrefixedWriter)(nil)

type LogWriter struct {
	w  io.Writer
	mu sync.Mutex
	wg sync.WaitGroup
}

// Write implements io.Writer.
func (s *LogWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

var _ io.Writer = (*LogWriter)(nil)

func (s *LogWriter) WithPrefix(prefix string) io.Writer {
	if prefix != "" && hasANSISupport() {
		prefix = types.GetStyledPrefix(prefix)
	}

	return &PrefixedWriter{
		w:      s.w,
		prefix: []byte(prefix),
		buf:    bytes.NewBuffer(nil),
		render: func(b []byte) []byte { return b },
	}
}

func (s *LogWriter) WithDimmedPrefix(prefix string) io.Writer {
	if prefix != "" && hasANSISupport() {
		prefix = types.GetDimStyledPrefix(prefix)
	}

	return &PrefixedWriter{
		w:      s.w,
		prefix: []byte(prefix),
		buf:    bytes.NewBuffer(nil),
		render: func(b []byte) []byte { return []byte(types.GetDimmedText(b)) },
	}
}
