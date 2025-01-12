package handlers

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/auth"
	"code-garden-server/utils"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"net/http"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(db *database.DBClient) *AuthHandler {
	s := auth.NewAuthService(db)
	h := &AuthHandler{s}
	return h
}

type requestBody struct {
	Email string `json:"email"`
}

func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	tokenString := r.PathValue("token")
	if tokenString == "" {
		utils.WriteRes(w, utils.Response{Status: 400, Message: "Bad request: Invalid token", Error: "Invalid token"})
		return
	}

	err := h.service.VerifyToken(tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteRes(w, utils.Response{Status: 404, Message: "failed to verify token", Error: err.Error()})
		} else {
			utils.WriteRes(w, utils.Response{Status: 500, Message: "failed to verify token", Error: err.Error()})
		}

		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Verification completed"})
}

func (h *AuthHandler) RegisterWithEmail(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 400, Error: err.Error(), Message: "Bad request body"})
		return
	}

	err = h.service.RegisterWithEmail(body.Email, "http://localhost:3000/auth/verify-token")
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 500, Error: err.Error(), Message: "Failed to create account"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Success! Please check your mail to verify your account"})
}

func (h *AuthHandler) LoginWithEmail(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 400, Error: err.Error(), Message: "Bad request body"})
		return
	}

	err = h.service.LoginWithEmail(body.Email, "http://localhost:8080/auth/verify-token")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteRes(w, utils.Response{Status: 404, Error: err.Error(), Message: "User with email not found."})
		} else {
			utils.WriteRes(w, utils.Response{Status: 400, Error: err.Error(), Message: "Failed to send email"})
		}
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Email sent"})
}
