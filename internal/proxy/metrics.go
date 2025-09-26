package proxy

import (
	"sort"
	"sync"
	"time"
)

type snapshot struct {
	Samples int           `json:"samples"`
	QPS     float64       `json:"qps"`
	P50     time.Duration `json:"p50"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	Errors  int           `json:"errors"`
}

type Metrics struct {
	mu     sync.Mutex
	lat    []time.Duration
	errors int
	start  time.Time
}

func NewMetrics() *Metrics { return &Metrics{start: time.Now()} }

func (m *Metrics) Record(d time.Duration, err error) {
	m.mu.Lock()
	m.lat = append(m.lat, d)
	if err != nil {
		m.errors++
	}
	m.mu.Unlock()
}

func (m *Metrics) Snapshot() snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.lat) == 0 {
		return snapshot{Samples: 0, QPS: 0}
	}
	ls := append([]time.Duration(nil), m.lat...)
	sort.Slice(ls, func(i, j int) bool { return ls[i] < ls[j] })

	p := func(q float64) time.Duration {
		idx := int(float64(len(ls)-1) * q)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(ls) {
			idx = len(ls) - 1
		}
		return ls[idx]
	}

	secs := time.Since(m.start).Seconds()
	qps := float64(len(m.lat)) / secs

	return snapshot{
		Samples: len(m.lat),
		QPS:     qps,
		P50:     p(0.50),
		P95:     p(0.95),
		P99:     p(0.99),
		Errors:  m.errors,
	}
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lat = nil
	m.errors = 0
	m.start = time.Now()
}
