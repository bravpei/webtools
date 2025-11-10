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

var defaultLogger *slog.Logger

func InitLogs() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := &CustomHandler{
		handler: slog.NewTextHandler(os.Stdout, opts),
	}
	defaultLogger = slog.New(handler)
}

func SetLogLevel(level slog.Level) {
	if defaultLogger == nil {
		InitLogs()
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := &CustomHandler{
		handler: slog.NewTextHandler(os.Stdout, opts),
	}
	defaultLogger = slog.New(handler)
}

type CustomHandler struct {
	handler slog.Handler
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
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

	// 构造日志格式
	logMsg := fmt.Sprintf(
		"%s [%s] [%s:%d] %s\n",
		timestamp,
		level,
		file,
		line,
		r.Message,
	)

	fmt.Print(logMsg)
	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		handler: h.handler.WithGroup(name),
	}
}

func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		InitLogs()
	}
	return defaultLogger
}
