package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/onkelwolle/chirpy/internal/database"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Cannot connect to database: %s", err)
	}

	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		templates: loadTemplates(),
		dbQueries: database.New(db),
	}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validate_chirp)
	mux.HandleFunc("GET /api/healthz", healthz)

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

func loadTemplates() *template.Template {
	tmpl, err := template.ParseFiles("templates/admin_metrics.html")
	if err != nil {
		log.Println("Error loading templates:", err)
	}
	return tmpl
}
