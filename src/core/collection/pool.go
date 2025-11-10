package collection

import (
	"context"
	"sync"
)

// BoundedPool limits concurrent goroutine execution
// Mathematical guarantee: Never exceeds N concurrent workers
type BoundedPool struct {
	workers   int
	semaphore chan struct{}
	wg        sync.WaitGroup
}

// NewBoundedPool creates a pool with N maximum workers
// Complexity: O(1)
func NewBoundedPool(maxWorkers int) *BoundedPool {
	return &BoundedPool{
		workers:   maxWorkers,
		semaphore: make(chan struct{}, maxWorkers),
	}
}

// Submit adds a task to the pool
// Blocks if pool is full (backpressure)
// Complexity: O(1)
func (p *BoundedPool) Submit(ctx context.Context, task func()) error {
	// Acquire semaphore slot
	select {
	case p.semaphore <- struct{}{}:
		// Slot acquired
	case <-ctx.Done():
		return ctx.Err()
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.semaphore }() // Release slot

		task()
	}()

	return nil
}

// Wait blocks until all tasks complete
// Complexity: O(1)
func (p *BoundedPool) Wait() {
	p.wg.Wait()
}
