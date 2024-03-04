package fpool

import (
	"errors"
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
	cap         int32         // 最大worker数量
	running     int32         // 正在运行的worker数量
	workers     []*Worker     // 空闲的worker
	expire      time.Duration // worker过期时间
	release     chan sig      // 释放worker的信号
	lock        sync.Mutex    // 保护workers的锁
	once        sync.Once     // 保证只释放一次
	workerCache sync.Pool     // 缓存
	cond        *sync.Cond
	panicHandle func()
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
		workers: make([]*Worker, 0),
		expire:  expire,
		release: make(chan sig, 1),
	}
	pool.workerCache.New = func() any {
		return &Worker{
			pool: pool,
			task: make(chan func(), 1),
		}
	}
	pool.cond = sync.NewCond(&pool.lock)
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
	return nil
}

// GetWorker 获取一个worker
func (p *Pool) GetWorker() *Worker {
	p.lock.Lock()
	// 如果有空闲的worker，直接获取
	idleWorker := p.workers
	n := len(idleWorker)
	if n > 0 {
		w := idleWorker[n-1]
		p.workers = idleWorker[:n-1]
		p.lock.Unlock()
		return w
	}
	p.lock.Unlock()

	// 如果容量没超过限制且没有空闲的worker，创建一个新的worker
	if p.running < p.cap {
		c := p.workerCache.Get()
		var w *Worker
		if c == nil {
			w = &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
		} else {
			w = c.(*Worker)
		}
		// worker 创建后启动一个携程持续监听任务
		w.run()
		return w
	}
	// 如果正在运行的worker 大于最大worker数量，阻塞等待 worker 释放
	return p.waitWorker()
}

func (p *Pool) waitWorker() *Worker {
	p.lock.Lock()
	defer p.lock.Unlock()

	//fmt.Println("cond 等待通知")
	p.cond.Wait()
	//fmt.Println("cond 得到通知，有空闲的worker")

	idleWorker := p.workers
	n := len(idleWorker) - 1
	// 如果没有空间，阻塞等待
	if n < 0 {
		return p.waitWorker()
	}
	w := idleWorker[n]
	p.workers = idleWorker[:n]
	return w
}

// Put 将worker放回池中
func (p *Pool) Put(w *Worker) {
	w.lastTime = time.Now()

	p.lock.Lock()
	defer p.lock.Unlock()

	p.workers = append(p.workers, w)

	p.cond.Signal()
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// descRunning 减少正在运行的worker数量
func (p *Pool) descRunning(c int32) {
	atomic.AddInt32(&p.running, -c)
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
	go p.expireWorkerTicker()
	return true
}

// expireWorkerTicker 清理过期 worker
func (p *Pool) expireWorkerTicker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			return
		}
		p.expireWorker()
	}
}

func (p *Pool) expireWorker() {
	p.lock.Lock()
	defer p.lock.Unlock()
	idleWorker := p.workers
	n := len(idleWorker)
	if n == 0 {
		return
	}
	var clearN = -1
	now := time.Now()
	for index, w := range idleWorker {
		// 遇到第一个没有过期的worker，停止清理
		if now.Sub(w.lastTime) <= p.expire {
			break
		}
		clearN = index
		w.task <- nil
	}
	if clearN == -1 {
		return
	}
	//fmt.Println("清除过期worker开始 ", p.running, len(idleWorker))
	if clearN == len(idleWorker)-1 {
		// 如果最后一个过期的worker在末尾，说明前面的worker已经全部过期，清空操作
		//fmt.Println("清除过期worker全部过期", "正在运行", p.running, "过期", len(idleWorker))
		p.descRunning(int32(len(idleWorker)))
		p.workers = idleWorker[:0]
		for _, w := range idleWorker {
			p.workerCache.Put(w)
		}
	} else {
		//fmt.Println("清除过期worker部分过期", p.running, "空闲", len(idleWorker), "过期", len(idleWorker[:clearN]))
		p.workers = idleWorker[clearN+1:]
		for _, w := range idleWorker[0 : clearN+1] {
			p.workerCache.Put(w)
		}
		p.descRunning(int32(len(idleWorker[0 : clearN+1])))
	}
	//fmt.Println("清除过期worker完成", "正在运行", p.running, "空闲", len(p.workers))
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}

func (p *Pool) GetIdleWorkerCount() any {
	return len(p.workers)
}
