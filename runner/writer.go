package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
)

type LineWriter struct {
	w io.Writer
}

// Write implements io.Writer.
func (lw *LineWriter) Write(p []byte) (n int, err error) {
	return lw.w.Write(p)
}

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
	pr, pw := io.Pipe()
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		copyStream(prefix, s.w, pr)
	}()
	return &LineWriter{w: pw}
}

func (s *LogWriter) Wait() {
	s.wg.Wait()
}

func copyStream(prefix string, dest io.Writer, src io.Reader) {
	r := bufio.NewReader(src)
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if prefix != "" {
					dest.Write([]byte(fmt.Sprintf("[%s] ", prefix)))
				}
				dest.Write(b)
				return
			}
		}

		if prefix != "" {
			dest.Write([]byte(fmt.Sprintf("[%s] ", prefix)))
		}
		dest.Write(b)
	}
}
