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
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

func (cfg *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}
	type returnVals struct {
		Chirp
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

	respondWithJSON(w, http.StatusCreated, returnVals{
		Chirp{
			Id:        chi.ID.String(),
			CreatedAt: chi.CreatedAt.String(),
			UpdatedAt: chi.UpdatedAt.String(),
			Body:      chi.Body,
			UserId:    chi.UserID.String(),
		},
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
