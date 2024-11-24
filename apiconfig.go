package main

import (
	"html/template"
	"sync/atomic"

	"github.com/onkelwolle/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	templates      *template.Template
	dbQueries      *database.Queries
	secret         []byte
}
