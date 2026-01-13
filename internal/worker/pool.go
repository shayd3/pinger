package worker

import (
	"context"
	"sync"

	"github.com/shayd3/pinger/internal/checker"
	"github.com/shayd3/pinger/internal/config"
)

type Job struct {
	Target config.Target
}

type Pool struct {
	workers    int
	jobs       chan Job
	results    chan checker.Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewPool(workers int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workers:    workers,
		jobs:       make(chan Job, workers*2),
		results:    make(chan checker.Result, workers*2),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			c := checker.New(job.Target.Type)

			checkCtx, cancel := context.WithTimeout(p.ctx, job.Target.Timeout)
			result := c.Check(checkCtx, job.Target)
			cancel()

			select {
			case p.results <- result:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

func (p *Pool) Submit(job Job) {
	select {
	case p.jobs <- job:
	case <-p.ctx.Done():
	}
}

func (p *Pool) Results() <-chan checker.Result {
	return p.results
}

func (p *Pool) Stop() {
	p.cancelFunc()
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}

func Run(ctx context.Context, targets []config.Target, concurrency int) []checker.Result {
	pool := NewPool(concurrency)
	pool.Start()

	var results []checker.Result
	var resultsMu sync.Mutex
	done := make(chan struct{})

	go func() {
		for result := range pool.Results() {
			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()
		}
		close(done)
	}()

	for _, target := range targets {
		pool.Submit(Job{Target: target})
	}

	close(pool.jobs)
	pool.wg.Wait()
	close(pool.results)

	<-done

	return results
}
