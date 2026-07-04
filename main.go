package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./"))
	appHandler := http.StripPrefix("/app/", fileServer)
	cfg := &apiConfig{}
	wrappedAppHandler := cfg.middlewareMetricsInc(appHandler)
	mux.Handle("/app/", wrappedAppHandler)
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /reset", cfg.resetHandler)
	s := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("Serving on port %s\n", port)
	log.Fatal(s.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
