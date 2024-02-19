package logger

import (
	"encoding/json"
	"time"
)

type JsonFormatter struct {
	TimeDisplay bool
}

func (f *JsonFormatter) Format(params *FormatterParams) string {
	now := time.Now().Format("2006-01-02 15:04:05")

	if params.Fields == nil {
		params.Fields = make(Fields)
	}
	if f.TimeDisplay {
		params.Fields["time"] = now
	}
	params.Fields["msg"] = params.Msg
	params.Fields["level"] = params.Level.Level()
	marshal, err := json.Marshal(params.Fields)
	if err != nil {
		panic(err)
	}
	return string(marshal) + "\n"
}
