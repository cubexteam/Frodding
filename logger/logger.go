package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorWhite  = "\033[97m"
)

type Logger struct {
	debug bool
	out   io.Writer
	err   io.Writer
}

func New(debug bool) *Logger {
	return &Logger{debug: debug, out: os.Stdout, err: os.Stderr}
}

func (l *Logger) timestamp() string {
	return colorGray + time.Now().Format("15:04:05") + colorReset
}

func (l *Logger) log(w io.Writer, level, color, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "%s %s%s%s %s\n", l.timestamp(), color, level, colorReset, msg)
}

func (l *Logger) Infof(format string, args ...any)  { l.log(l.out, "INFO ", colorCyan, format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.log(l.out, "WARN ", colorYellow, format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.log(l.err, "ERROR", colorRed, format, args...) }

func (l *Logger) Debugf(format string, args ...any) {
	if l.debug {
		l.log(l.out, "DEBUG", colorGray, format, args...)
	}
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.log(l.err, "FATAL", colorRed, format, args...)
	os.Exit(1)
}

func (l *Logger) Chat(name, message string) {
	fmt.Fprintf(l.out, "%s %sCHAT %s<%s> %s\n", l.timestamp(), colorWhite, colorReset, name, message)
}
