package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	defaultLogger *Logger
)

const (
	DefaultLoggerFlag = log.Ldate | log.Ltime | log.Lmicroseconds
)

func init() {
	defaultLogger = New(os.Stdout, "", DefaultLoggerFlag, LogLevelInfo)
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
	case LogLevelError.String():
		return LogLevelError, nil
	case LogLevelWarn.String():
		return LogLevelWarn, nil
	case LogLevelInfo.String():
		return LogLevelInfo, nil
	case LogLevelDebug.String():
		return LogLevelDebug, nil
	case LogLevelTrace.String():
		return LogLevelTrace, nil
	default:
		return LogLevelError, fmt.Errorf("unknown log level: %s", level)
	}
}

func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

type Logger struct {
	logger *log.Logger
	level  LogLevel
}

func New(out io.Writer, prefix string, flag int, level LogLevel) *Logger {
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
		s := fmt.Sprintf("[%s] %s", level, fmt.Sprintf(format, args...))
		l.logger.Print(s)
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
