package repo_pool_async

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type RepoUpdatePoolBusyWaiting struct {
	q               *Queue
	dynamicPriority bool
	muLock          *sync.Mutex
	//condProducer     *sync.Cond
	condConsumer     *sync.Cond
	availableWorkers uint32
	logger           *zap.Logger
}

func NewPool(ctx context.Context, maxSizeWorker int, logger *zap.Logger) *RepoUpdatePoolBusyWaiting {
	p := &RepoUpdatePoolBusyWaiting{
		q:      &Queue{},
		muLock: &sync.Mutex{},
		logger: logger,
	}

	//p.condProducer = sync.NewCond(p.muLock)
	p.condConsumer = sync.NewCond(p.muLock)

	for i := 0; i < maxSizeWorker; i++ {
		go p.runWorker(ctx)
	}

	return p
}

type Handler func(ctx context.Context)
type Job struct {
	handler      Handler
	expiredTime  int64
	responseTime int64
}

func (j *Job) process(ctx context.Context) {
	if j.responseTime >= j.expiredTime {
		return
	}
	j.handler(ctx)
}

func (j *Job) Run(ctx context.Context) {
	j.responseTime = time.Now().UnixMilli()
}

func (p *RepoUpdatePoolBusyWaiting) PushPriority(ctx context.Context, handler Handler) *Job {
	// priority true mean if don't have any
	// workers available then use strategy priority
	// spawn temporary worker
	// to process fastest
	// good for acid transaction call cache update
	job := &Job{
		handler:      handler,
		expiredTime:  time.Now().Add(100 * time.Millisecond).UnixMilli(),
		responseTime: time.Now().Add(100 * time.Millisecond).UnixMilli(),
	}
	if atomic.LoadUint32(&p.availableWorkers) == 0 {
		go func() {
			job.process(ctx)
		}()
	} else {
		p.q.En(job)
		fmt.Println("repo_pool_async not empty: ", p.q.Empty())
		p.condConsumer.Signal()
	}

	return job
}

func (p *RepoUpdatePoolBusyWaiting) runWorker(ctx context.Context) {
	atomic.AddUint32(&p.availableWorkers, 1)
	for {
		// lock to check cond
		p.muLock.Lock()
		for p.q.Empty() {
			p.condConsumer.Wait()
		}
		// if can wake up because have job then
		// pop front job but no release mutex lock for sync
		job := p.q.De().(*Job)
		p.muLock.Unlock()
		time.Sleep(time.Millisecond * 300)
		// now release lock

		p.logger.Info("Process Job")
		// when receive job then decrease available workers
		atomic.AddUint32(&p.availableWorkers, ^uint32(0))

		// process job with response time smaller than expired time
		job.process(ctx)
		// when worker finish job then increase available workers
		atomic.AddUint32(&p.availableWorkers, 1)
	}
}
