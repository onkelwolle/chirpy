package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		templates: loadTemplates(),
	}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

	mux.HandleFunc("GET /api/healthz", healthz)

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("Server listening on port 8080...")
	err := srv.ListenAndServe()
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
