package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ktothebro/InferenceProxy/internal/proxy"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	workers := flag.Int("workers", 32, "worker count")
	enableBatch := flag.Bool("batch", false, "enable batching")
	batchSize := flag.Int("batchSize", 8, "max batch size")
	batchFlushMs := flag.Int("batchFlushMs", 10, "flush timer ms")
	timeoutMs := flag.Int("timeoutMs", 400, "per request timeout ms")
	retries := flag.Int("retries", 0, "retry attempts")
	flag.Parse()

	s := proxy.NewServer(proxy.Config{
		Workers:      *workers,
		EnableBatch:  *enableBatch,
		BatchSize:    *batchSize,
		BatchFlushMs: *batchFlushMs,
		TimeoutMs:    *timeoutMs,
		Retries:      *retries,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/infer", s.HandleInfer)
	mux.HandleFunc("/metrics", s.HandleMetrics)
	mux.HandleFunc("/metrics/reset", s.HandleReset)

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
