package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/onkelwolle/chirpy/internal/config"
	"github.com/onkelwolle/chirpy/internal/database"
	"github.com/onkelwolle/chirpy/internal/handler"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Cannot connect to database: %s", err)
	}

	mux := http.NewServeMux()

	apiCfg := &config.ApiConfig{
		Templates:             loadTemplates(),
		DbQueries:             database.New(db),
		Secret:                []byte(os.Getenv("SECRET")),
		AccessTokenExpiresIn:  60 * 60 * 1,       // 1 hour
		RefreshTokenExpiresIn: 60 * 60 * 24 * 60, // 60 days
	}

	fileServer := http.FileServer(http.Dir("."))
	configureEndpoints(mux, apiCfg, fileServer)

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("Server listening on port 8080...")
	err = srv.ListenAndServe()
	if err != nil {
		log.Println("Error starting server:", err)
	}
}

func configureEndpoints(mux *http.ServeMux, apiCfg *config.ApiConfig, fileServer http.Handler) {

	chirpHandler := handler.NewChirpHandler(apiCfg)
	metricsHandler := handler.NewMetricsHandler(apiCfg)
	userHandler := handler.NewUsersHandler(apiCfg)

	mux.Handle("/app/", metricsHandler.MiddlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	mux.HandleFunc("GET /admin/metrics", metricsHandler.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", metricsHandler.ResetMetricsHandler)

	mux.HandleFunc("POST /api/chirps", chirpHandler.CreateChirps)
	mux.HandleFunc("GET /api/chirps", chirpHandler.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", chirpHandler.GetChirpByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", chirpHandler.DeleteChirp)

	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("POST /api/login", userHandler.Login)
	mux.HandleFunc("PUT /api/users", userHandler.UpdateUser)
	mux.HandleFunc("POST /api/refresh", userHandler.RefreshToken)
	mux.HandleFunc("POST /api/revoke", userHandler.RevokeToken)

	mux.HandleFunc("GET /api/healthz", healthz)
}

func loadTemplates() *template.Template {
	tmpl, err := template.ParseFiles("templates/admin_metrics.html")
	if err != nil {
		log.Println("Error loading templates:", err)
	}
	return tmpl
}
