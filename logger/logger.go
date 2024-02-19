package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type Level int
type Fields = map[string]interface{}

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

type Logger struct {
	Level       Level
	Formatter   IFormatter
	Out         []*LoggerWriter
	Fields      Fields
	LogPathDir  string
	LogFileSize int64
}

func New() *Logger {
	return &Logger{}
}

type LoggerWriter struct {
	Level Level
	W     io.Writer
}

type IFormatter interface {
	Format(params *FormatterParams) string
}

type FormatterParams struct {
	Level         Level
	IsColored     bool
	Fields        Fields
	Msg           string
	KeysAndValues []any
}

type Formatter struct {
	Level     Level
	IsColored bool
	Fields    Fields
}

func (l *Level) Level() string {
	switch *l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

func Default() *Logger {
	w := &LoggerWriter{
		Level: LevelDebug,
		W:     os.Stdout,
	}
	return &Logger{
		Level:     LevelDebug,
		Out:       []*LoggerWriter{w}, // 默认输出到控制台
		Formatter: &JsonFormatter{},
	}
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.Print(LevelInfo, msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.Print(LevelError, msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.Print(LevelDebug, msg, keysAndValues...)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	l.Fields = fields
	return l
}

func (l *Logger) SetLevel(level Level) {
	l.Level = level
}

func (l *Logger) Print(level Level, msg string, keysAndValues ...any) {
	// 如果日志级别小于设置的级别， 则不输出
	if level < l.Level {
		return
	}
	params := &FormatterParams{
		Level:         level,
		Fields:        l.Fields,
		Msg:           msg,
		KeysAndValues: keysAndValues,
	}
	for _, out := range l.Out {
		if out.W == os.Stdout {
			params.IsColored = true
			fmt.Fprint(out.W, l.Formatter.Format(params))
		} else if out.Level == -1 || out.Level == level {
			l.CheckFileSize(out)
			fmt.Fprint(out.W, l.Formatter.Format(params))
		}
	}
}

func (l *Logger) SetLoggerPath(logPathDir string) {
	l.LogPathDir = logPathDir
	_, err := os.Stat(logPathDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(logPathDir, os.ModePerm)
		}
	}
	l.Out = append(l.Out, &LoggerWriter{
		Level: -1,
		W:     FileWriter(path.Join(logPathDir, "all.log")),
	})
	l.Out = append(l.Out, &LoggerWriter{
		Level: LevelDebug,
		W:     FileWriter(path.Join(logPathDir, "debug.log")),
	})
	l.Out = append(l.Out, &LoggerWriter{
		Level: LevelInfo,
		W:     FileWriter(path.Join(logPathDir, "info.log")),
	})
	l.Out = append(l.Out, &LoggerWriter{
		Level: LevelError,
		W:     FileWriter(path.Join(logPathDir, "error.log")),
	})
}

func (l *Logger) CheckFileSize(out *LoggerWriter) {
	logFile := out.W.(*os.File)
	if logFile == nil {
		return
	}
	// 判断文件大小
	fileInfo, err := logFile.Stat()
	if err != nil {
		panic(err)
	}
	if l.LogFileSize <= 0 {
		l.LogFileSize = 100 << 20 // 100M
	}
	if fileInfo.Size() > l.LogFileSize {
		// 文件超过大小，重新创建文件
		_, name := path.Split(fileInfo.Name())
		logFile.Close()
		err = os.Rename(path.Join(l.LogPathDir, fileInfo.Name()), path.Join(l.LogPathDir, fmt.Sprintf("%s_%d.log", strings.TrimSuffix(name, ".log"), time.Now().Unix())))
		if err != nil {
			panic(err)
		}
		newFile := FileWriter(path.Join(l.LogPathDir, name))
		out.W = newFile
	}

}

func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return w
}
