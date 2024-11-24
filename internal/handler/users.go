package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/onkelwolle/chirpy/internal/auth"
	"github.com/onkelwolle/chirpy/internal/config"
	"github.com/onkelwolle/chirpy/internal/database"
	"github.com/onkelwolle/chirpy/internal/models"
	"github.com/onkelwolle/chirpy/internal/utils"
)

type usersHandler struct {
	cfg *config.ApiConfig
}

func NewUsersHandler(cfg *config.ApiConfig) *usersHandler {
	return &usersHandler{cfg: cfg}
}

func (u *usersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := u.cfg.DbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.User{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	})

}

func (u *usersHandler) Login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := u.cfg.DbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	err = auth.ComparePassword(user.HashedPassword, params.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	var expiresIn int
	if params.ExpiresInSeconds == 0 || params.ExpiresInSeconds > 3600 {
		expiresIn = 3600
	} else {
		expiresIn = params.ExpiresInSeconds
	}

	token, err := auth.MakeJWT(user.ID, string(u.cfg.Secret), time.Duration(expiresIn)*time.Second)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.User{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		Token:     token,
	})

}
