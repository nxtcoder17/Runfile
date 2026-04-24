package writer

import (
	"bytes"
	"errors"
	"io"
	"sync"
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
			pw.w.Write(pw.render(line))
			break
		}

		if _, err := pw.w.Write(pw.prefix); err != nil {
			return n, err
		}

		if _, err := pw.w.Write(pw.render(line)); err != nil {
			return n, err
		}
	}
	return n, nil
}

var _ io.Writer = (*PrefixedWriter)(nil)

type LogWriter struct {
	io.Writer
	Mu sync.Mutex
	wg sync.WaitGroup
}

// Write implements io.Writer.
func (s *LogWriter) Write(p []byte) (n int, err error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Writer.Write(p)
}

var _ io.Writer = (*LogWriter)(nil)

func (s *LogWriter) WithPrefix(prefix string) io.Writer {
	if prefix != "" {
		if IsANSITerminal() {
			prefix = GetStyledPrefix(prefix) + " "
		} else {
			prefix = "[" + prefix + "] "
		}
	}

	return &PrefixedWriter{
		w:      s.Writer,
		prefix: []byte(prefix),
		buf:    bytes.NewBuffer(nil),
		render: func(b []byte) []byte { return b },
	}
}

func (s *LogWriter) WithDimmedPrefix(prefix string) io.Writer {
	var render func([]byte) []byte
	if prefix != "" {
		if IsANSITerminal() {
			prefix = GetDimStyledPrefix(prefix) + " "
			render = func(b []byte) []byte { return []byte(GetDimmedText(b)) }
		} else {
			prefix = "[" + prefix + "] "
			render = func(b []byte) []byte { return b }
		}
	} else {
		render = func(b []byte) []byte { return b }
	}

	return &PrefixedWriter{
		w:      s.Writer,
		prefix: []byte(prefix),
		buf:    bytes.NewBuffer(nil),
		render: render,
	}
}
