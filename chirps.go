package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/onkelwolle/chirpy/internal/auth"
	"github.com/onkelwolle/chirpy/internal/database"
	"github.com/onkelwolle/chirpy/internal/models"
)

func (cfg *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	userId, err := auth.ValidateJWT(bearerToken, string(cfg.secret))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(chirp.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chi, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody(chirp.Body),
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, models.Chirp{
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

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {

	dbChirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps", err)
		return
	}

	chirps := convertDatabaseChirps(dbChirps)
	respondWithJSON(w, http.StatusOK, chirps)
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

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("chirpId")
	log.Printf("Getting chirp with ID: %s", id)
	chirpID, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid chirp ID", err)
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, models.Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt.String(),
		UpdatedAt: dbChirp.UpdatedAt.String(),
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	})
}
