package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/onkelwolle/chirpy/internal/config"
	"github.com/onkelwolle/chirpy/internal/utils"
)

type metricsHandler struct {
	cfg *config.ApiConfig
}

func NewMetricsHandler(cfg *config.ApiConfig) *metricsHandler {
	return &metricsHandler{cfg: cfg}
}

func (m *metricsHandler) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (m *metricsHandler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := struct {
		Hits int32
	}{
		Hits: m.cfg.FileserverHits.Load(),
	}
	if err := m.cfg.Templates.ExecuteTemplate(w, "admin_metrics.html", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (m *metricsHandler) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		utils.RespondWithError(w, http.StatusForbidden, "Reset not allowed", nil)
		return
	}
	err := m.cfg.DbQueries.DeleteUsers(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not reset users", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	oldHits := m.cfg.FileserverHits.Swap(0) // Atomically reset the counter
	fmt.Fprintf(w, "Resetting metrics. Previous count: %d, Resetting Database State to initial state.\n", oldHits)
}
