package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func createLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := &CustomHandler{
		level:   level,
		handler: slog.NewTextHandler(os.Stdout, opts),
	}
	logger := slog.New(handler)
	// 设置为全局默认logger
	slog.SetDefault(logger)
	return logger
}

func InitLogs() {
	createLogger(slog.LevelInfo)
}

func SetLogLevel(level slog.Level) {
	createLogger(level)
}

type CustomHandler struct {
	level    slog.Level
	handler  slog.Handler
	preAttrs []slog.Attr
	group    string
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// 时间格式：YYYY-MM-DD HH:MM:SS
	timestamp := r.Time.Format(time.DateTime)

	// 日志级别（大写）
	level := strings.ToUpper(r.Level.String())

	// 获取调用者信息（文件名和行号）
	file := ""
	line := 0
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			// 只取文件名，不显示完整路径
			file = filepath.Base(f.File)
			line = f.Line
		}
	}

	// 构造基本日志格式
	logMsg := fmt.Sprintf(
		"%s [%s] [%s:%d] %s",
		timestamp,
		level,
		file,
		line,
		r.Message,
	)

	// 添加预挂载属性（来自 WithAttrs）
	for _, attr := range h.preAttrs {
		logMsg += formatAttr(attr)
	}

	// 添加每次调用的属性
	r.Attrs(func(attr slog.Attr) bool {
		logMsg += formatAttr(attr)
		return true
	})

	logMsg += "\n"
	fmt.Print(logMsg)
	return nil
}

func formatAttr(attr slog.Attr) string {
	switch attr.Value.Kind() {
	case slog.KindString:
		return fmt.Sprintf(" %s: %s", attr.Key, attr.Value.String())
	case slog.KindInt64:
		return fmt.Sprintf(" %s: %d", attr.Key, attr.Value.Int64())
	case slog.KindFloat64:
		return fmt.Sprintf(" %s: %f", attr.Key, attr.Value.Float64())
	case slog.KindBool:
		return fmt.Sprintf(" %s: %t", attr.Key, attr.Value.Bool())
	case slog.KindAny:
		if err, ok := attr.Value.Any().(error); ok {
			return fmt.Sprintf(" %s: %s", attr.Key, err.Error())
		}
		return fmt.Sprintf(" %s: %v", attr.Key, attr.Value.Any())
	default:
		return fmt.Sprintf(" %s: %v", attr.Key, attr.Value)
	}
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.preAttrs)+len(attrs))
	copy(newAttrs, h.preAttrs)
	copy(newAttrs[len(h.preAttrs):], attrs)

	newHandler := *h
	newHandler.preAttrs = newAttrs
	return &newHandler
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	if h.group != "" {
		newHandler.group = h.group + "." + name
	} else {
		newHandler.group = name
	}
	return &newHandler
}
