package main

import (
	"database/sql"
	"os"

	"github.com/jabeztay/bootdev-chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	JWTSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	const port = "8080"
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./"))
	appHandler := http.StripPrefix("/app/", fileServer)
	cfg := &apiConfig{}
	cfg.dbQueries = database.New(db)
	cfg.platform = os.Getenv("PLATFORM")
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	wrappedAppHandler := cfg.middlewareMetricsInc(appHandler)
	mux.Handle("/app/", wrappedAppHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", cfg.chirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.postUsersHandler)
	mux.HandleFunc("POST /api/login", cfg.loginUserHandler)
	mux.HandleFunc("POST /api/refresh", cfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", cfg.revokeRefreshTokenHandler)
	s := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("Serving on port %s\n", port)
	log.Fatal(s.ListenAndServe())
}
