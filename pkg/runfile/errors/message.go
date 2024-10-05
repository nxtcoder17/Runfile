package errors

import (
	"encoding/json"
	"log/slog"
)

type Message interface {
	error

	WithMetadata(attrs ...any) *message
	Log()
}

type message struct {
	text string
	err  error

	metadata []any
}

func New(text string, err error) *message {
	return &message{
		text: text,
		err:  err,
	}
}

func (m *message) WithErr(err error) *message {
	m.err = err
	return m
}

func (m *message) WithMetadata(metaAttrs ...any) *message {
	maxlen := len(metaAttrs)
	if len(metaAttrs)&1 == 1 {
		// INFO: if odd, leave last item
		maxlen -= 1
	}

	for i := 0; i < maxlen; i += 2 {
		m.metadata = append(m.metadata, metaAttrs[i], metaAttrs[i+1])
	}

	return m
}

func (m *message) Error() string {
	b, err := json.Marshal(map[string]any{
		"text":     m.text,
		"error":    m.err.Error(),
		"metadata": m.metadata,
	})
	if err != nil {
		panic(err)
	}

	return string(b)
}

func (m *message) Log() {
	if m.err == nil {
		slog.Error(m.text)
		return
	}
	slog.Error(m.text, "err", m.err)
}
