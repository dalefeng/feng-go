package logger

import (
	"fmt"
	"io"
	"os"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

type Logger struct {
	Level     Level
	Formatter *Formatter
	Out       []io.Writer
}

func New() *Logger {
	return &Logger{}
}

type Formatter struct {
}

func Default() *Logger {
	return &Logger{
		Level:     LevelDebug,
		Out:       append([]io.Writer{}, os.Stdout), // 默认输出到控制台
		Formatter: &Formatter{},
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.Print(LevelInfo, msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {

}

func (l *Logger) Debug(msg string, args ...interface{}) {

}

func (l *Logger) Print(level Level, msg string, args ...any) {
	// 如果日志级别小于设置的级别，则不输出
	if level < l.Level {
		return
	}
	for _, out := range l.Out {
		fmt.Fprint(out, msg)
	}
}

func (l *Logger) SetFormatter(formatter *Formatter) {
	l.Formatter = formatter
}

func (l *Logger) SetOutput(out io.Writer) {
	l.Out = append(l.Out, out)
}

func (l *Logger) SetLevel(level Level) {
	l.Level = level
}
