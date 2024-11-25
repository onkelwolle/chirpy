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
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt.String(),
		UpdatedAt:   user.UpdatedAt.String(),
		IsChirpyRed: user.IsChirpyRed,
		Email:       user.Email,
	})

}

func (u *usersHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	token, err := auth.MakeJWT(user.ID, string(u.cfg.Secret), time.Duration(u.cfg.AccessTokenExpiresIn)*time.Second)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	_, err = u.cfg.DbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(u.cfg.RefreshTokenExpiresIn) * time.Second),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.User{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.String(),
		UpdatedAt:    user.UpdatedAt.String(),
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	})

}

func (u *usersHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	refreshTokenData, err := u.cfg.DbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	if refreshTokenData.ExpiresAt.Before(time.Now()) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Token expired", nil)
		return
	}

	if refreshTokenData.Revoked.Valid {
		utils.RespondWithError(w, http.StatusUnauthorized, "Token revoked", nil)
		return
	}

	user, err := u.cfg.DbQueries.GetUserByID(r.Context(), refreshTokenData.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't get user", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, string(u.cfg.Secret), time.Duration(u.cfg.AccessTokenExpiresIn)*time.Second)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func (u *usersHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	err = u.cfg.DbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke token", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusNoContent, struct{}{})
}

func (u *usersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, string(u.cfg.Secret))
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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

	user, err := u.cfg.DbQueries.UpdateUsersPasswordAndEmail(r.Context(), database.UpdateUsersPasswordAndEmailParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.User{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt.String(),
		UpdatedAt:   user.UpdatedAt.String(),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})

}
