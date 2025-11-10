package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func InitLogs() {
	log.SetReportCaller(true)
	log.SetFormatter(&CustomFormatter{})
}
func SetLogLevel(level uint32) {
	log.SetLevel(log.Level(level))
}

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	// 时间格式：YYYY-MM-DD HH:MM:SS
	timestamp := entry.Time.Format(time.DateTime)

	// 日志级别（大写）
	level := strings.ToUpper(entry.Level.String())

	// 获取调用者信息（文件名和行号）
	file := ""
	line := 0
	if entry.HasCaller() {
		// 只取文件名，不显示完整路径
		file = filepath.Base(entry.Caller.File)
		line = entry.Caller.Line
	}

	// 构造日志格式
	logMsg := fmt.Sprintf(
		"%s [%s] [%s:%d] %s\n",
		timestamp,
		level,
		file,
		line,
		entry.Message,
	)

	return []byte(logMsg), nil
}
func GetLogger() *log.Logger {
	return log.StandardLogger()
}
