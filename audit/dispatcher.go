package audit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryPublisher is a simple in-memory implementation of Publisher.
// It now uses a bounded buffer for storage and a channel for distribution.
type MemoryPublisher struct {
	mu     sync.RWMutex
	events []event
	maxSize int
	ch     chan event
}

func NewMemoryPublisher(bufferSize int) *MemoryPublisher {
	return &MemoryPublisher{
		ch:      make(chan event, bufferSize),
		events:  make([]event, 0, bufferSize),
		maxSize: bufferSize,
	}
}

func (p *MemoryPublisher) Publish(ctx context.Context, evt event) error {
	select {
	case p.ch <- evt:
		p.mu.Lock()
		defer p.mu.Unlock()
		
		// Fix Memory Leak: Use a circular buffer approach for storage.
		if len(p.events) >= p.maxSize {
			p.events = p.events[1:] // Simple FIFO eviction if full
		}
		p.events = append(p.events, evt)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("publisher buffer full")
	}
}

// DefaultDispatcher implements the Dispatcher interface with worker pooling, batching, and refined retry logic.
type DefaultDispatcher struct {
	publisher *MemoryPublisher
	sinks     []Sink
	stopCh    chan struct{}
	wg        sync.WaitGroup

	retryInterval time.Duration
	maxRetries    int
	
	// Worker pool settings
	workerCount int
	
	// Batching settings
	batchSize     int
	batchTimeout  time.Duration
}

func NewDefaultDispatcher(pub *MemoryPublisher, sinks []Sink) *DefaultDispatcher {
	return &DefaultDispatcher{
		publisher:     pub,
		sinks:         sinks,
		stopCh:        make(chan struct{}),
		retryInterval: 1 * time.Second,
		maxRetries:    3,
		workerCount:   5,   // Default worker pool size
		batchSize:     10,  // Default batch size
		batchTimeout:  2 * time.Second,
	}
}

func (d *DefaultDispatcher) Start(ctx context.Context) error {
	for i := 0; i < d.workerCount; i++ {
		d.wg.Add(1)
		go d.worker()
	}
	return nil
}

func (d *DefaultDispatcher) Stop(ctx context.Context) error {
	close(d.stopCh)
	
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *DefaultDispatcher) worker() {
	defer d.wg.Done()

	var batch []event
	ticker := time.NewTicker(d.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case evt, ok := <-d.publisher.ch:
			if !ok {
				return
			}
			batch = append(batch, evt)
			if len(batch) >= d.batchSize {
				d.dispatchBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				d.dispatchBatch(batch)
				batch = nil
			}
		case <-d.stopCh:
			// Drain remaining batch before stopping
			if len(batch) > 0 {
				d.dispatchBatch(batch)
			}
			return
		}
	}
}

func (d *DefaultDispatcher) dispatchBatch(batch []event) {
	for _, sink := range d.sinks {
		d.wg.Add(1)
		go func(s Sink, b []event) {
			defer d.wg.Done()
			
			backoff := d.retryInterval
			for i := 0; i <= d.maxRetries; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				var err error
				
				// Batching for external sinks
				if batchingSink, ok := s.(BatchingSink); ok {
					err = batchingSink.SendBatch(ctx, b)
				} else {
					// Fallback to individual sends if sink doesn't support batching
					for _, e := range b {
						if err = s.Send(ctx, e); err != nil {
							break
						}
					}
				}
				cancel()

				if err == nil {
					return
				}

				// Refine retry logic: distinguish between transient and permanent errors
				if !IsTransientError(err) {
					// Permanent error (e.g., 4xx), don't retry
					return
				}

				if i < d.maxRetries {
					select {
					case <-time.After(backoff):
						backoff *= 2
					case <-d.stopCh:
						return
					}
				}
			}
		}(sink, batch)
	}
}
