package proxy

import (
	"context"
	"time"
)

// Start a concurrent batcher: an aggregator builds batches and hands them off
// to a pool of batch processors (size = cfg.Workers) so multiple batches
// can run in parallel.
func startBatcher(in <-chan *job, cfg Config, m *Metrics) {
	// pool limits parallel batch processing
	sem := make(chan struct{}, max(1, cfg.Workers)) // reuse Workers as batch-concurrency
	// channel to hand off full batches to processors
	batchC := make(chan []*job, 1024)

	// processors
	for i := 0; i < cap(sem); i++ {
		go func() {
			for b := range batchC {
				sem <- struct{}{}
				start := time.Now()
				body, err := doWithRetries(b[0].ctx, cfg.Retries, func(ctx context.Context) ([]byte, error) {
					return simulateModel(ctx, len(b))
				})
				elapsed := time.Since(start)
				for _, j := range b {
					m.Record(elapsed, err)
					j.respCh <- response{body: body, err: err}
				}
				<-sem
			}
		}()
	}

	// aggregator: builds batches, flushes by size or timer
	go func() {
		defer close(batchC)
		buf := make([]*job, 0, cfg.BatchSize)
		t := time.NewTicker(time.Duration(cfg.BatchFlushMs) * time.Millisecond)
		defer t.Stop()

		flush := func() {
			if len(buf) == 0 {
				return
			}
			// hand off a copy so we can continue aggregating immediately
			b := make([]*job, len(buf))
			copy(b, buf)
			buf = buf[:0]
			batchC <- b
		}

		for {
			select {
			case j, ok := <-in:
				if !ok {
					flush()
					return
				}
				buf = append(buf, j)
				if len(buf) >= cfg.BatchSize {
					flush()
				}
			case <-t.C:
				flush()
			}
		}
	}()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
