package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type Config struct {
	Workers, BatchSize, BatchFlushMs, TimeoutMs, Retries int
	EnableBatch                                          bool
}

type Server struct {
	cfg     Config
	submitC chan *job
	metrics *Metrics
}

type job struct {
	ctx     context.Context
	payload []byte
	respCh  chan response
}

type response struct {
	body []byte
	err  error
}

func NewServer(cfg Config) *Server {
	s := &Server{
		cfg:     cfg,
		submitC: make(chan *job, 10000),
		metrics: NewMetrics(),
	}
	// one or the other
	if cfg.EnableBatch {
		go startBatcher(s.submitC, cfg, s.metrics)
	} else {
		go startWorkers(s.submitC, cfg, s.metrics)
	}
	return s
}

func (s *Server) HandleInfer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payload, _ := io.ReadAll(r.Body)

	// timeout per request
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.cfg.TimeoutMs)*time.Millisecond)
	defer cancel()

	j := &job{ctx: ctx, payload: payload, respCh: make(chan response, 1)}

	// quick drop if queue full
	select {
	case s.submitC <- j:
	default:
		http.Error(w, "queue_full", http.StatusServiceUnavailable)
		return
	}

	res := <-j.respCh
	if res.err != nil {
		http.Error(w, res.err.Error(), http.StatusGatewayTimeout)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res.body)
}

func (s *Server) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	m := s.metrics.Snapshot()
	b, _ := json.MarshalIndent(m, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) HandleReset(w http.ResponseWriter, r *http.Request) {
	s.metrics.Reset()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("reset\n"))
}

func simulateModel(ctx context.Context, batchSize int) ([]byte, error) {
	// Fixed overhead + small per-batch cost; amortized across the whole batch
	baseMs := 70 + rand.Intn(31)    // 70–100ms core work
	extraMs := 3 + rand.Intn(3)     // 3–5ms per batch overhead (not per item)

	total := time.Duration(baseMs+extraMs) * time.Millisecond
	select {
	case <-time.After(total):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if rand.Intn(100) < 2 {
		return nil, fmt.Errorf("upstream_error")
	}
	out := map[string]any{
		"ok":              true,
		"batch_processed": batchSize,
		"latency_ms":      baseMs + extraMs,
	}
	b, _ := json.Marshal(out)
	return b, nil
}
