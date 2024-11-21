package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	apiCfg := &apiConfig{}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	mux.HandleFunc("GET /api/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.resetMetricsHandler)

	mux.HandleFunc("GET /api/healthz", healthz)

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Println("Server listening on port 8080...")
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
