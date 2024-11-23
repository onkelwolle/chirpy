package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

func (cfg *apiConfig) validate_chirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	cleaned_body := cleanBody(chirp.Body)

	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: cleaned_body,
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
