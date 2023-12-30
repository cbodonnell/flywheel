package log

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

func init() {
	once.Do(func() {
		defaultLogger = New(os.Stdout, "", log.Ldate|log.Ltime, LogLevelDebug)
	})
}

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

func (level LogLevel) String() string {
	switch level {
	case LogLevelError:
		return "error"
	case LogLevelWarn:
		return "warn"
	case LogLevelInfo:
		return "info"
	case LogLevelDebug:
		return "debug"
	case LogLevelTrace:
		return "trace"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a log level string into a LogLevel.
// Valid log levels are: error, warn, info, debug, trace.
func ParseLogLevel(level string) (LogLevel, error) {
	switch level {
	case "error":
		return LogLevelError, nil
	case "warn":
		return LogLevelWarn, nil
	case "info":
		return LogLevelInfo, nil
	case "debug":
		return LogLevelDebug, nil
	case "trace":
		return LogLevelTrace, nil
	default:
		return LogLevelError, fmt.Errorf("unknown log level: %s", level)
	}
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
	defaultLogger.Info("Log level set to %s", level)
}

type Logger struct {
	logger *log.Logger
	level  LogLevel
}

func New(out *os.File, prefix string, flag int, level LogLevel) *Logger {
	return &Logger{
		logger: log.New(out, prefix, flag),
		level:  level,
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if level <= l.level {
		logEntry := map[string]interface{}{
			"level": level.String(),
			"msg":   fmt.Sprintf(format, args...),
		}
		msgBytes, _ := json.Marshal(logEntry)
		l.logger.Print(string(msgBytes))
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(LogLevelError, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.logf(LogLevelWarn, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(LogLevelInfo, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(LogLevelDebug, format, args...)
}

func (l *Logger) Trace(format string, args ...interface{}) {
	l.logf(LogLevelTrace, format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Trace(format string, args ...interface{}) {
	defaultLogger.Trace(format, args...)
}
