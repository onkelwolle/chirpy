package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := struct {
		Hits int32
	}{
		Hits: cfg.fileserverHits.Load(),
	}
	if err := cfg.templates.ExecuteTemplate(w, "admin_metrics.html", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	oldHits := cfg.fileserverHits.Swap(0) // Atomically reset the counter
	fmt.Fprintf(w, "Resetting metrics. Previous count: %d\n", oldHits)
}
