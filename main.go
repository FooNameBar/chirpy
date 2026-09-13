package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	apiCfg := apiConfig{fileserveHits: atomic.Int32{}}

	mux.Handle("GET /app/", middlewareLog(apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))))
	mux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK\n"))
	}))
	mux.Handle("GET /metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /reset", http.HandlerFunc(apiCfg.handlerReset))
	server.ListenAndServe()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
