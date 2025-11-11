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
	// 设置为全局默认logger
	slog.SetDefault(defaultLogger)
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
	// 更新全局默认logger
	slog.SetDefault(defaultLogger)
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

	// 构造基本日志格式
	logMsg := fmt.Sprintf(
		"%s [%s] [%s:%d] %s",
		timestamp,
		level,
		file,
		line,
		r.Message,
	)

	// 统一处理所有日志属性（包括错误信息）
	r.Attrs(func(attr slog.Attr) bool {
		switch attr.Value.Kind() {
		case slog.KindString:
			logMsg += fmt.Sprintf(" %s: %s", attr.Key, attr.Value.String())
		case slog.KindInt64:
			logMsg += fmt.Sprintf(" %s: %d", attr.Key, attr.Value.Int64())
		case slog.KindFloat64:
			logMsg += fmt.Sprintf(" %s: %f", attr.Key, attr.Value.Float64())
		case slog.KindBool:
			logMsg += fmt.Sprintf(" %s: %t", attr.Key, attr.Value.Bool())
		case slog.KindAny:
			// 处理错误类型和其他任意类型
			if err, ok := attr.Value.Any().(error); ok {
				logMsg += fmt.Sprintf(" %s: %s", attr.Key, err.Error())
			} else {
				logMsg += fmt.Sprintf(" %s: %v", attr.Key, attr.Value.Any())
			}
		default:
			logMsg += fmt.Sprintf(" %s: %v", attr.Key, attr.Value)
		}
		return true
	})

	logMsg += "\n"
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
