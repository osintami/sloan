// Copyright © 2023 TrailerCamz
// Copyright © 2025 Sloan Kendall Childers III
package log

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	LOG_TRACE = 0xFF
	LOG_FATAL = 0x00
	LOG_INFO  = 0x04
	LOG_WARN  = 0x02
	LOG_ERROR = 0x01
)

type ILogger interface {
	Str(string, string) ILogger
	Int(string, int) ILogger
	Int64(string, int64) ILogger
	Float(string, float32) ILogger
	Bool(string, bool) ILogger
	Err(error) ILogger
	Msg(string)
}

type Logger struct {
	parts  []string
	ignore bool
}

// NOTE global logging variables
var LOG_FH *os.File
var LOG_FILE string
var LOG_LEVEL int = LOG_ERROR
var LOG_STDERR bool = true

func InitLogger(path, file, level string, standardError bool) {

	LOG_FILE = fmt.Sprintf("%s/%s", path, file)
	LOG_STDERR = standardError

	switch strings.ToLower(level) {
	case "trace":
		LOG_LEVEL = LOG_TRACE
	case "debug":
		LOG_LEVEL = LOG_TRACE
	case "error":
		LOG_LEVEL = LOG_ERROR
	case "warn":
		LOG_LEVEL = LOG_ERROR | LOG_WARN
	case "info":
		LOG_LEVEL = LOG_ERROR | LOG_WARN | LOG_INFO
	}

	if file != "" {
		if _, err := os.Stat(path); err != nil {
			os.Mkdir(path, 0700)
		}
		if _, err := os.Stat(LOG_FILE); err != nil {
			LOG_FH, err = os.Create(LOG_FILE)
			if err != nil {
				fmt.Println("[ERROR] create log file failed", err)
			}
		} else {
			LOG_FH, err = os.OpenFile(LOG_FILE, os.O_RDWR|os.O_APPEND, 0700)
			if err != nil {
				fmt.Println("[ERROR] open log file failed", err)
			}
		}
	}

	Info().Str("component", "osintami").Str("level", level).Str("file", LOG_FILE).Msg("logging started")
}

func LogFile() string {
	return LOG_FILE
}

func NewLogger(level int) *Logger {
	x := &Logger{parts: []string{}}
	if LOG_LEVEL&level != level {
		x.ignore = true
		return x
	}
	return x
}

func Info() ILogger {
	x := NewLogger(LOG_INFO)
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, "\"level\":\"info\"")
	return x
}

func Warn() ILogger {
	x := NewLogger(LOG_WARN)
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, "\"level\":\"warn\"")
	return x
}

func Error() ILogger {
	x := NewLogger(LOG_ERROR)
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, "\"level\":\"error\"")
	return x
}

func Fatal() ILogger {
	x := &Logger{parts: []string{}}
	x.parts = append(x.parts, "\"level\":\"fatal\"")
	return x
}

func Debug() ILogger {
	x := NewLogger(LOG_TRACE)
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, "\"level\":\"debug\"")
	return x
}

// TODO:  preserve stacktrace from one back
func (x *Logger) Err(err error) ILogger {
	if x.ignore || err == nil {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"error\":\"%s\"", err.Error()))
	return x
}

func (x *Logger) Str(key, value string) ILogger {
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"%s\":\"%s\"", key, value))
	return x
}

func (x *Logger) Bool(key string, value bool) ILogger {
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"%s\":\"%t\"", key, value))
	return x
}

func (x *Logger) Int(key string, value int) ILogger {
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"%s\":\"%d\"", key, value))
	return x
}

func (x *Logger) Int64(key string, value int64) ILogger {
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"%s\":\"%d\"", key, value))
	return x
}

func (x *Logger) Float(key string, value float32) ILogger {
	if x.ignore {
		return x
	}
	x.parts = append(x.parts, fmt.Sprintf("\"%s\":\"%f\"", key, value))
	return x
}

func (x *Logger) Msg(msg string) {
	if x.ignore {
		return
	}

	var buffer bytes.Buffer
	buffer.Write([]byte("{"))
	buffer.Write([]byte(fmt.Sprintf("\"time\":\"%s\",", time.Now().Format(time.RFC3339))))
	for _, item := range x.parts {
		buffer.Write([]byte(item))
		buffer.Write([]byte(","))
	}
	buffer.Write([]byte(fmt.Sprintf("\"message\":\"%s\"", msg)))
	buffer.Write([]byte("}"))
	buffer.Write([]byte("\n"))

	out := buffer.Bytes()
	if LOG_STDERR {
		os.Stderr.Write(out)
	}
	if LOG_FH != nil {
		LOG_FH.Write(out)
	}
}

func Shutdown() {
	LOG_FH.Close()
}
