package fesgo

import (
	"errors"
	"fmt"
	"github.com/dalefeng/fesgo/ferror"
	"runtime"
	"strings"
)

func Recovery(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger.Error(detailMsg(err))
				originErr := err.(error)
				if originErr != nil {
					var ferr *ferror.FesError
					if errors.As(originErr, &ferr) {
						ferr.ExecuteResult()
						return
					}
				}
				ctx.Abort(errors.New("internal server error"))
			}
		}()
		next(ctx)
	}
}

func detailMsg(err any) string {
	// 上溯的栈帧数，O表示Caller的调用者（Caller所在的调用栈））（0-当前函数，1-上一层函数，）。
	//runtime.Caller(3)

	// pc 调用栈标识符
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n %v \n", err))
	for _, pc := range pcs[0:n] {
		// 获取到对应的函数
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n \t %s:%d", file, line))
	}

	return sb.String()
}
