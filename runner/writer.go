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
		copyStreamLineByLine(prefix, s.w, pr)
	}()
	return &LineWriter{w: pw}
}

func (s *LogWriter) Wait() {
	s.wg.Wait()
}

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Green = "\033[32m"
)

func copyStreamLineByLine(prefix string, dest io.Writer, src io.Reader) {
	hasPrefix := prefix != ""
	if hasPrefix && hasANSISupport() {
		prefix = fmt.Sprintf("%s[%s]%s ", Green, prefix, Reset)
		// prefix = fmt.Sprintf("%s%s |%s ", Green, prefix, Reset)
	}
	r := bufio.NewReader(src)
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if hasPrefix {
					dest.Write([]byte(prefix))
				}
				dest.Write(b)
				return
			}
		}

		if hasPrefix {
			dest.Write([]byte(prefix))
		}
		dest.Write(b)
	}
}
