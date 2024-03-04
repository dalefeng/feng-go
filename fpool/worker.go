package fpool

import (
	fesLog "github.com/dalefeng/fesgo/logger"
	"time"
)

type Worker struct {
	pool     *Pool
	task     chan func()
	lastTime time.Time // 上次执行任务的时间
}

func (w *Worker) run() {
	w.pool.incRunning()
	go w.running()
}

// running 持久运行任务
func (w *Worker) running() {
	defer func() {
		if err := recover(); err != nil {
			if w.pool.panicHandle != nil {
				w.pool.panicHandle()
			} else {
				fesLog.Default().Error("running panic", err)
			}
			// 重新启动 running
			fesLog.Default().Info("panic restart running task")
			w.run()
			w.pool.Put(w)
		}
	}()
	for f := range w.task {
		// 任务为空时，退出
		if f == nil {
			w.pool.cond.Signal()
			return
		}
		f()
		// 运行完任务后，将worker放回池中
		w.pool.Put(w)
	}
}
