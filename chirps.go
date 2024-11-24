package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/onkelwolle/chirpy/internal/database"
)

type Chirp struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserID    string `json:"user_id"`
}

func (cfg *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := parameters{}
	err := decoder.Decode(&chirp)
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

	userId, err := uuid.Parse(chirp.UserId)
	chi, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody(chirp.Body),
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
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

func convertDatabaseChirps(dbChirps []database.Chirp) []Chirp {
	chirps := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
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
	type returnVals struct {
		Chirp
	}

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

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt.String(),
		UpdatedAt: dbChirp.UpdatedAt.String(),
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	})
}
