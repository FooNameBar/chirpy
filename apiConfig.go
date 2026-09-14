package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/FooNameBar/chirpy/internal/database"
)

type apiConfig struct {
	fileserveHits atomic.Int32
	db            *database.Queries
	platform      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserveHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	template := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`
	fmt.Fprintf(w, template, cfg.fileserveHits.Load())
}
