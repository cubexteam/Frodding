package logger

import (
	"fmt"
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
}

func New(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (l *Logger) timestamp() string {
	return colorGray + time.Now().Format("15:04:05") + colorReset
}

func (l *Logger) Infof(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %sINFO%s  %s\n", l.timestamp(), colorCyan, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %sWARN%s  %s\n", l.timestamp(), colorYellow, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s %sERROR%s %s\n", l.timestamp(), colorRed, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	if !l.debug {
		return
	}
	fmt.Fprintf(os.Stdout, "%s %sDEBUG%s %s\n", l.timestamp(), colorGray, colorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s %sFATAL%s %s\n", l.timestamp(), colorRed, colorReset, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *Logger) Chat(name, message string) {
	fmt.Fprintf(os.Stdout, "%s %sCHAT%s  <%s> %s\n", l.timestamp(), colorWhite, colorReset, name, message)
}
