package errors

import (
	"encoding/json"
	"log/slog"
)

type Message struct {
	text string
	err  error

	metadataKeys map[string]int
	metadata     []any
}

var _ error = (*Message)(nil)

func New(text string, err error) *Message {
	return &Message{text: text, err: err}
}

func (m Message) WithErr(err error) *Message {
	m.err = err
	return &m
}

func (m Message) WithMetadata(metaAttrs ...any) *Message {
	if m.metadataKeys == nil {
		m.metadataKeys = make(map[string]int)
	}

	m.metadata = append(m.metadata, metaAttrs...)

	return &m
}

func (m *Message) Error() string {
	m2 := map[string]any{
		"text":     m.text,
		"metadata": m.metadata,
	}
	if m.err != nil {
		m2["error"] = m.err.Error()
	}

	b, err := json.Marshal(m2)
	if err != nil {
		slog.Error("marshalling, got", "err", err)
		panic(err)
	}

	return string(b)
}

func (m *Message) Log() {
	if m.err == nil {
		slog.Error(m.text, m.metadata...)
		return
	}
	// args := []any{"err", m.err}
	// args = append(args, m.metadata...)
	// slog.Error(m.text, args...)
	slog.Error(m.err.Error(), m.metadata...)
}
