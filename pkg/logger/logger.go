// pkg/logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Level int

const (
	InfoLevel Level = iota
	WarnLevel
	ErrorLevel
	DebugLevel
)

var (
	logger *log.Logger
	level  Level
)

// Init initializes the logger with optional file output
func Init(logLevel Level, logFile string) error {
	level = logLevel

	flags := log.Ldate | log.Ltime | log.LUTC

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logger = log.New(file, "", flags)
	} else {
		logger = log.New(os.Stdout, "", flags)
	}

	return nil
}

func formatMessage(lvl Level, message string) string {
	_, file, line, _ := runtime.Caller(2)
	fileName := filepath.Base(file)
	levelStr := "INFO"
	switch lvl {
	case WarnLevel:
		levelStr = "WARN"
	case ErrorLevel:
		levelStr = "ERROR"
	case DebugLevel:
		levelStr = "DEBUG"
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s [%s] %s:%d: %s", timestamp, levelStr, fileName, line, message)
}

func Info(format string, v ...interface{}) {
	if level >= InfoLevel {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatMessage(InfoLevel, msg))
	}
}

func Warn(format string, v ...interface{}) {
	if level >= WarnLevel {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatMessage(WarnLevel, msg))
	}
}

func Error(format string, v ...interface{}) {
	if level >= ErrorLevel {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatMessage(ErrorLevel, msg))
	}
}

func Debug(format string, v ...interface{}) {
	if level >= DebugLevel {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatMessage(DebugLevel, msg))
	}
}

// Fatal logs a message and exits the program
func Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.Fatalln(formatMessage(ErrorLevel, msg))
}
