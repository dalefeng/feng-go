package fpool

import "time"

type Worker struct {
	pool     *Pool
	task     chan func()
	lastTime time.Time // 上次执行任务的时间
}

func (w *Worker) run() {
	go w.running()
}

// running 持久运行任务
func (w *Worker) running() {
	for f := range w.task {
		if f != nil {
			f()
		}
		// 运行完任务后，将worker放回池中
		w.pool.Put(w)
		w.pool.descRunning()
	}
}
