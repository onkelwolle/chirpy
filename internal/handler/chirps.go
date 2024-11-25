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
	var dbChirps []database.Chirp
	var err error

	sort := "asc"
	if r.URL.Query().Get("sort") == "desc" {
		sort = "desc"
	}

	authorId := r.URL.Query().Get("author_id")
	if authorId != "" {
		authorUUID, err := uuid.Parse(authorId)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}

		if sort == "desc" {
			dbChirps, err = h.cfg.DbQueries.GetChirpsByUserIDDesc(r.Context(), authorUUID)
		} else {
			dbChirps, err = h.cfg.DbQueries.GetChirpsByUserIDAsc(r.Context(), authorUUID)
		}
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get chirps", err)
			return
		}

		chirps := convertDatabaseChirps(dbChirps)
		utils.RespondWithJSON(w, http.StatusOK, chirps)
		return
	}

	if sort == "desc" {
		dbChirps, err = h.cfg.DbQueries.GetChirpsDesc(r.Context())

	} else {
		dbChirps, err = h.cfg.DbQueries.GetChirpsAsc(r.Context())
	}
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
		utils.RespondWithError(w, http.StatusNotFound, "Could not get chirp", err)
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

func (h *chirpHandler) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, string(h.cfg.Secret))
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	id := r.PathValue("chirpId")
	chirpID, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Invalid chirp ID", err)
		return
	}

	chirp, err := h.cfg.DbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Could not get chirp", err)
		return
	}

	if chirp.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp", nil)
		return
	}

	log.Printf("Deleting chirp with ID: %s", id)
	err = h.cfg.DbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not delete chirp", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusNoContent, nil)
}
