package fesgo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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

var DefaultWriter io.Writer = os.Stdout

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LoggerFormatter = func(params *LogFormatterParams) string

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	isDisplayColor bool
}

func (p LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return cyan
	case code >= 400 && code < 500:
		return red
	default:
		return red
	}
}

func (p LogFormatterParams) MethodColor() string {
	switch p.Method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

var defaultFormatter = func(param *LogFormatterParams) string {

	// 超过分钟转换为秒
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	if param.isDisplayColor {
		return fmt.Sprintf("%s[fesgo]%s %s %v %s |%s %3d %s| %s %10s %s | %13v |%s %-7s %s  %s %#v %s \n",
			yellow, reset,
			blue, param.TimeStamp.Format("2006-01-02 15:04:05"), reset,
			param.StatusCodeColor(), param.StatusCode, reset,
			red, param.Latency, reset,
			param.ClientIP,
			param.MethodColor(), param.Method, reset,
			cyan, param.Path, reset)
	}
	return fmt.Sprintf("[fesgo] %s | %3d | %10s | %13v | %-7s %#v \n",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method, param.Path)
}

func LoggingWithConfig(config LoggingConfig, next HandlerFunc) HandlerFunc {
	formatter := config.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}

	out := config.out

	displayColor := false

	if out == nil {
		out = DefaultWriter
	}
	if out == DefaultWriter {
		displayColor = true
	}

	return func(c *Context) {

		r := c.R
		start := time.Now()
		next(c)
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		clientIP := net.ParseIP(ip)

		path := r.URL.Path
		raw := r.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		param := &LogFormatterParams{
			Request:        c.R,
			TimeStamp:      stop,
			StatusCode:     c.StatusCode,
			Latency:        latency,
			ClientIP:       clientIP,
			Method:         r.Method,
			Path:           path,
			isDisplayColor: displayColor,
		}
		fmt.Fprint(out, formatter(param))
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
