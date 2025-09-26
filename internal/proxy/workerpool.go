package proxy

import (
	"context"
	"time"
)

func startWorkers(in <-chan *job, cfg Config, m *Metrics) {
	for i := 0; i < cfg.Workers; i++ {
		go func() {
			for j := range in {
				start := time.Now()
				body, err := doWithRetries(j.ctx, cfg.Retries, func(ctx context.Context) ([]byte, error) {
					return simulateModel(ctx, 1)
				})
				m.Record(time.Since(start), err)
				j.respCh <- response{body: body, err: err}
			}
		}()
	}
}

func doWithRetries(ctx context.Context, retries int, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	var body []byte
	var err error
	for a := 0; a <= retries; a++ {
		body, err = fn(ctx)
		if err == nil {
			return body, nil
		}
		// simple backoff unless context is done
		select {
		case <-time.After(time.Duration(20*(a+1)) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, err
}
