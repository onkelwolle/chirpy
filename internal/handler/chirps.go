package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/onkelwolle/chirpy/internal/auth"
	"github.com/onkelwolle/chirpy/internal/config"
	"github.com/onkelwolle/chirpy/internal/database"
	"github.com/onkelwolle/chirpy/internal/models"
	utils "github.com/onkelwolle/chirpy/internal/utils"
)

type chirpHandler struct {
	cfg *config.ApiConfig
}

func NewChirpHandler(cfg *config.ApiConfig) *chirpHandler {
	return &chirpHandler{cfg: cfg}
}

func (h *chirpHandler) CreateChirps(w http.ResponseWriter, r *http.Request) {

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	userId, err := auth.ValidateJWT(bearerToken, string(h.cfg.Secret))
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := parameters{}
	err = decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(chirp.Body) > maxChirpLength {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chi, err := h.cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody(chirp.Body),
		UserID: userId,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.Chirp{
		ID:        chi.ID.String(),
		CreatedAt: chi.CreatedAt.String(),
		UpdatedAt: chi.UpdatedAt.String(),
		Body:      chi.Body,
		UserID:    chi.UserID.String(),
	})
}

func cleanBody(body string) string {
	targetWords := []string{"kerfuffle", "sharbert", "fornax"}

	// Replace each word (case-insensitively) with "****"
	for _, word := range targetWords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		body = re.ReplaceAllString(body, "****")
	}

	return body
}

func (h *chirpHandler) GetChirps(w http.ResponseWriter, r *http.Request) {

	dbChirps, err := h.cfg.DbQueries.GetChirps(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not get chirps", err)
		return
	}

	chirps := convertDatabaseChirps(dbChirps)
	utils.RespondWithJSON(w, http.StatusOK, chirps)
}

func convertDatabaseChirps(dbChirps []database.Chirp) []models.Chirp {
	chirps := make([]models.Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = models.Chirp{
			ID:        dbChirp.ID.String(),
			CreatedAt: dbChirp.CreatedAt.String(),
			UpdatedAt: dbChirp.UpdatedAt.String(),
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID.String(),
		}
	}
	return chirps
}

func (h *chirpHandler) GetChirpByID(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("chirpId")
	log.Printf("Getting chirp with ID: %s", id)
	chirpID, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Invalid chirp ID", err)
		return
	}

	dbChirp, err := h.cfg.DbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not get chirp", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt.String(),
		UpdatedAt: dbChirp.UpdatedAt.String(),
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	})
}
