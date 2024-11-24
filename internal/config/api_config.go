package config

import (
	"html/template"
	"sync/atomic"

	"github.com/onkelwolle/chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Templates      *template.Template
	DbQueries      *database.Queries
	Secret         []byte
}
