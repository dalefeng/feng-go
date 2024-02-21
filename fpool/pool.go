package fpool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 3 * time.Second

var (
	ErrorInvalidCap    = errors.New("cap must be greater than 0")
	ErrorInvalidExpire = errors.New("expire must be greater than 0")
	ErrorHasClosed     = errors.New("pool has bean released")
)

type sig struct {
}

type Pool struct {
	cap     int32         // 最大worker数量
	running int32         // 正在运行的worker数量
	workers []*Worker     // 空闲的worker
	expire  time.Duration // worker过期时间
	release chan sig      // 释放worker的信号
	lock    sync.Mutex    // 保护workers的锁
	once    sync.Once     // 保证只释放一次
}

func NewPool(cap int32) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int32, expire time.Duration) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorInvalidCap
	}
	pool := &Pool{
		cap:     cap,
		running: 0,
		workers: make([]*Worker, 0, cap),
		expire:  expire,
		release: make(chan sig, 1),
	}
	go pool.expireWorkerTicker()
	return pool, nil
}

// Submit 提交一个任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorHasClosed
	}
	// 获取池里面的worker，然后执行任务
	w := p.GetWorker()
	w.task <- task
	w.pool.incRunning()
	return nil
}

// GetWorker 获取一个worker
func (p *Pool) GetWorker() *Worker {
	p.lock.Lock()
	// 如果有空闲的worker，直接获取
	idleWorker := p.workers
	n := len(idleWorker) - 1
	if n > 0 {
		w := idleWorker[n]
		p.workers = idleWorker[:n]
		p.lock.Unlock()
		return w
	}
	// 如果容量没超过限制且没有空闲的worker，创建一个新的worker
	if p.running < p.cap {
		w := &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
		// worker 创建后启动一个携程持续监听任务
		w.run()
		p.lock.Unlock()
		return w
	}
	p.lock.Unlock()

	// 如果正在运行的worker 大于最大worker数量，阻塞等待 worker 释放
	for {
		p.lock.Lock()
		idleWorker = p.workers
		n = len(idleWorker) - 1
		// 如果没有空间，阻塞等待
		if n < 0 {
			p.lock.Unlock()
			continue
		}
		w := idleWorker[n]
		p.workers = idleWorker[:n]
		p.lock.Unlock()
		return w
	}
}

// Put 将worker放回池中
func (p *Pool) Put(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workers = append(p.workers, w)
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// descRunning 减少正在运行的worker数量
func (p *Pool) descRunning() {
	atomic.AddInt32(&p.running, -1)
}

// Release 释放池
func (p *Pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		for i, w := range p.workers {
			w.task = nil
			w.pool = nil
			p.workers[i] = nil
		}
		p.workers = nil
		p.release <- sig{}
	})
}

// IsClosed 池是否已经关闭
func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

// Restart 重启池
func (p *Pool) Restart() bool {
	if p.IsClosed() {
		return true
	}
	<-p.release
	p.workers = make([]*Worker, 0, 1)
	return true
}

// expireWorkerTicker 清理过期 worker
func (p *Pool) expireWorkerTicker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		fmt.Println("expireWorkerTicker", time.Now().Format("2006-01-02 15:04:05"))
		p.expireWorker()
	}
}

func (p *Pool) expireWorker() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.workers) == 0 {
		return
	}
	now := time.Now()
	for index, w := range p.workers {
		if now.Sub(w.lastTime) <= p.expire {
			continue
		}
		p.workers = append(p.workers[:index], p.workers[index+1:]...)
	}
	// TODO

}
