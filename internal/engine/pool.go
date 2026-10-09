package engine

import (
	"context"
	"sync"
	"sync/atomic"
)

// Task is a unit of work executable by the worker pool.
type Task[T any, R any] func(ctx context.Context, item T) R

// Pool manages concurrent task execution with rate limiting and cancellation.
type Pool[T any, R any] struct {
	concurrency int
	task        Task[T, R]
	processed   int64
	errors      int64
}

// NewPool creates a new worker pool.
func NewPool[T any, R any](concurrency int, task Task[T, R]) *Pool[T, R] {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &Pool[T, R]{
		concurrency: concurrency,
		task:        task,
	}
}

// Run executes the pool over an input slice and collects results.
func (p *Pool[T, R]) Run(ctx context.Context, items []T) []R {
	if len(items) == 0 {
		return nil
	}

	results := make([]R, len(items))
	itemCh := make(chan struct {
		index int
		item  T
	}, len(items))

	for i, item := range items {
		itemCh <- struct {
			index int
			item  T
		}{index: i, item: item}
	}
	close(itemCh)

	var wg sync.WaitGroup
	workers := p.concurrency
	if workers > len(items) {
		workers = len(items)
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range itemCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				res := p.task(ctx, work.item)
				results[work.index] = res
				atomic.AddInt64(&p.processed, 1)
			}
		}()
	}

	wg.Wait()
	return results
}

// ProcessedCount returns number of processed items.
func (p *Pool[T, R]) ProcessedCount() int64 {
	return atomic.LoadInt64(&p.processed)
}
