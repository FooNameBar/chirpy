package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/FooNameBar/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("godotenv.Load: %v\n", err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserveHits: atomic.Int32{},
		db:            dbQueries,
		platform:      os.Getenv("PLATFORM"),
		secret:        os.Getenv("JWT_SECRET"),
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.Handle("GET /app/", middlewareLog(apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))))

	mux.Handle("GET /api/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK\n"))
	}))
	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.handlerCreateUser))
	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.handlerLogin))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handlerCreateChirp))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.handleGetChirps))
	mux.Handle("GET /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.handleGetChirpByID))

	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.handlerReset))
	server.ListenAndServe()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
