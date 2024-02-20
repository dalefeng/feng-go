package ferror

// 优雅的错误处理
// 自定义error, 通过recover来处理

type FesError struct {
	err     error
	ErrFunc ErrorFunc
}

func Default() *FesError {
	return &FesError{}
}

func (e *FesError) Error() string {
	return e.err.Error()
}

func (e *FesError) Put(err error) {
	e.checkError(err)
}

func (e *FesError) checkError(err error) {
	if err == nil {
		return
	}
	e.err = err
	panic(e)
}

type ErrorFunc func(fesError *FesError)

func (e *FesError) Result(errFunc ErrorFunc) {
	e.ErrFunc = errFunc
}

func (e *FesError) ExecuteResult() {
	e.ErrFunc(e)
}
