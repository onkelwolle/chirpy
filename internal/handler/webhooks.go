package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/onkelwolle/chirpy/internal/auth"
	"github.com/onkelwolle/chirpy/internal/config"
	"github.com/onkelwolle/chirpy/internal/models"
	"github.com/onkelwolle/chirpy/internal/utils"
)

type webhooksHandler struct {
	cfg *config.ApiConfig
}

func NewWebhooksHandler(cfg *config.ApiConfig) *webhooksHandler {
	return &webhooksHandler{cfg: cfg}
}

func (wh *webhooksHandler) PolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid API key", err)
		return
	}

	if apiKey != string(wh.cfg.PolkaWebhookSecret) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	polka := models.Polka{}
	err = decoder.Decode(&polka)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if polka.Event != "user.upgraded" {
		utils.RespondWithError(w, http.StatusNoContent, "Invalid event", nil)
		return
	}

	userID, err := uuid.Parse(polka.Data.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	_, err = wh.cfg.DbQueries.UpdateUserToChirpyRed(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Couldn't update user", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusNoContent, nil)
}
