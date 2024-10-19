package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/phuslu/log"
)

type Options struct {
	Prefix string
	Writer io.Writer

	Theme *Theme

	ShowTimestamp bool
	ShowCaller    bool
	ShowDebugLogs bool
	ShowLogLevel  bool

	SetAsDefaultLogger bool

	// PickPrefix func() string
	SlogKeyAsPrefix string

	// constants
	keyValueSeparator string
}

func (opts Options) WithDefaults() Options {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	if opts.Theme == nil {
		opts.Theme = DefaultTheme()
	}

	if opts.keyValueSeparator == "" {
		opts.keyValueSeparator = "="
	}

	return opts
}

func New(opts Options) *slog.Logger {
	opts = opts.WithDefaults()

	writePrefix := func(w io.Writer, args *log.FormatterArgs) {
		if opts.Prefix != "" {
			fmt.Fprintf(w, "%s ", opts.Theme.TaskPrefixStyle.Render(opts.Prefix))
		}

		for i := range args.KeyValues {
			if args.KeyValues[i].Key == opts.SlogKeyAsPrefix {
				fmt.Fprintf(w, "%s ", opts.Theme.LogLevelStyles[ParseLogLevel(args.Level)].Render(fmt.Sprintf("[%s]", args.KeyValues[i].Value)))
				break
			}
		}
	}

	l := log.Logger{
		Level: func() log.Level {
			if opts.ShowDebugLogs {
				return log.DebugLevel
			}
			return log.InfoLevel
		}(),
		Caller: func() int {
			if opts.ShowCaller {
				return 1
			}
			return 0
		}(),
		// Context: []byte{},
		Writer: &log.ConsoleWriter{
			ColorOutput:    false,
			QuoteString:    true,
			EndWithMessage: true,
			Formatter: func(w io.Writer, args *log.FormatterArgs) (int, error) {
				writePrefix(w, args)

				if opts.ShowLogLevel {
					fmt.Fprintf(w, "%s ", opts.Theme.LogLevelStyles[ParseLogLevel(args.Level)].Render(args.Level))
				}

				fmt.Fprint(w, opts.Theme.MessageStyle.Render(args.Message))
				for i := range args.KeyValues {
					if args.KeyValues[i].Key == opts.SlogKeyAsPrefix {
						continue
					}
					fmt.Fprintf(w, " %s%s%v", opts.Theme.SlogKeyStyle.Render(args.KeyValues[i].Key), opts.Theme.SlogKeyStyle.Faint(true).Render(opts.keyValueSeparator), opts.Theme.MessageStyle.Render(args.KeyValues[i].Value))
				}

				return fmt.Fprintf(w, "\n")
			},
		},
	}
	sl := l.Slog()
	if opts.SetAsDefaultLogger {
		slog.SetDefault(sl)
	}
	return sl
}
