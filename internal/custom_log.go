package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

const (
	timeFormat  = "[15:04:05.000]"
	logFormat   = "%v\t%v\t%v"
	attrFormat  = "%v=%v"
	attrSep     = "\t"
	attrsFormat = "\t{%v}"
)

type CustomLogFileHandler struct {
	handler slog.Handler
	file    io.Writer
	buf     *bytes.Buffer
	mu      *sync.Mutex
}

func (h *CustomLogFileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *CustomLogFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomLogFileHandler{handler: h.handler.WithAttrs(attrs), file: h.file, buf: h.buf, mu: h.mu}
}

func (h *CustomLogFileHandler) WithGroup(name string) slog.Handler {
	return &CustomLogFileHandler{handler: h.handler.WithGroup(name), file: h.file, buf: h.buf, mu: h.mu}
}

func (h *CustomLogFileHandler) Handle(ctx context.Context, r slog.Record) error {
	// Fetch attributes
	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	attrLogs := ""
	for k, v := range attrs {
		if len(attrLogs) > 0 {
			attrLogs += attrSep
		}
		attrLogs += fmt.Sprintf(attrFormat, k, v)
	}

	logLine := fmt.Sprintf(logFormat, r.Time.Format(timeFormat), r.Level, r.Message)
	if len(attrLogs) > 0 {
		logLine += fmt.Sprintf(attrsFormat, attrLogs)
	}
	logLine += "\n"
	io.WriteString(h.file, logLine)

	return nil
}

func (h *CustomLogFileHandler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.mu.Lock()
	defer func() {
		h.buf.Reset()
		h.mu.Unlock()
	}()

	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling default handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buf.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when parsing default handle result: %w", err)
	}
	return attrs, nil
}

func NewCustomLogHandler(w io.Writer, opts *slog.HandlerOptions) *CustomLogFileHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	b := &bytes.Buffer{}
	return &CustomLogFileHandler{
		file: w,
		buf:  b,
		handler: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaultAttrs(opts.ReplaceAttr),
		}),
		mu: &sync.Mutex{},
	}
}

// Ignore the default log attributes
func suppressDefaultAttrs(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}
