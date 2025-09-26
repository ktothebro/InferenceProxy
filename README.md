# InferenceProxy (Go)

A tiny, production-style **inference proxy** written in Go. It simulates an upstream model call and focuses on **concurrency, batching, retries/timeouts, and observability** (QPS + p50/p95/p99). Great for demonstrating real-world infra trade-offs (throughput vs tail latency) in applied-ML/AI platform interviews.

---

## Features
- **Worker pool** for high concurrency
- **Concurrent batching** (batch size + time flush) to amortize upstream work
- **Retries + per-request timeouts** (simple backoff)
- **Built-in metrics** endpoint: QPS, p50, p95, p99, error count
- **Health check** and **metrics reset** endpoint for clean A/B runs

---

## How it works (simulated upstream)
Each request goes through the proxy. The “model” is simulated with a fixed cost (70–100ms) plus a small per-batch overhead. With batching, N requests share one upstream call, boosting throughput and usually improving tails (given the right timeouts).

---

## Run

```bash
# No batching (baseline)
go run ./cmd/proxy/main.go -addr=":8080" -workers=64 -batch=false -retries=0 -timeoutMs=1200

# Concurrent batching ON (tuned)
go run ./cmd/proxy/main.go -addr=":8080" -workers=192 -batch=true -batchSize=8 -batchFlushMs=3 -retries=1 -timeoutMs=2000
