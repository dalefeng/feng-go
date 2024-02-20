package logger

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(params *FormatterParams) string {
	now := time.Now().Format("2006-01-02 15:04:05")
	b := strings.Builder{}
	if params.IsColored {
		b.WriteString(fmt.Sprintf("%s[fesgo]%s %s %v %s | level=%s%s%s | msg=%s%v %s | ",
			yellow, reset,
			blue, now, reset,
			f.LevelColor(params.Level), params.Level.Level(), reset,
			f.MsgColor(params.Level), params.Msg, reset))
	} else {
		b.WriteString(fmt.Sprintf("[fesgo] %s | level=%s | msg=%v | ", now, params.Level.Level(), params.Msg))
	}
	fIndex := 0
	fLen := len(params.Fields)
	if params.Fields != nil {
		for k, v := range params.Fields {
			b.WriteString(fmt.Sprintf("%s=%#v", k, v))
			if fIndex < fLen-1 {
				b.WriteString(", ")
			}
			fIndex++
		}
	}
	kLen := len(params.Args)
	for index, arg := range params.Args {
		if index == 0 {
			b.WriteString(", ")
		}
		if params.IsKeysAndValues { // key=value格式
			i := index % 2
			switch arg.(type) {
			case nil:
				b.WriteString("nil")
			case string:
				if i == 1 {
					b.WriteString(fmt.Sprintf("%#v", arg))
				} else {
					b.WriteString(fmt.Sprintf("%s", arg))
				}
			default:
				// key
				if i == 0 {
					b.WriteString("key!%s")
				} else {
					b.WriteString(fmt.Sprintf("%+v", arg))
				}
			}
			if i == 0 && index < kLen-1 {
				b.WriteString("=")
			}
			if i == 1 && index < kLen-1 {
				b.WriteString(", ")
			}
		} else {
			switch arg.(type) {
			case nil:
				b.WriteString("nil")
			case string, error:
				b.WriteString(fmt.Sprintf("%s", arg))
			default:
				b.WriteString(fmt.Sprintf("%+v", arg))
			}
			if index < kLen-1 {
				b.WriteString(", ")
			}
		}

	}

	b.WriteString("\n")
	return b.String()
}

func (f *TextFormatter) LevelColor(level Level) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return reset
	}
}

func (f *TextFormatter) MsgColor(level Level) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return reset
	}
}
