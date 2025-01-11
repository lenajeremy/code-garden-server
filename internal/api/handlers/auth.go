package handlers

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/auth"
	"code-garden-server/utils"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	service *auth.AuthService
}

func NewAuthHandler(db *database.DBClient) *AuthHandler {
	s := auth.NewAuthService(db)
	h := &AuthHandler{s}
	return h
}

func (h *AuthHandler) LoginWithEmail(w http.ResponseWriter, r *http.Request) {
	type reqbody struct {
		Email string `json:"email"`
	}
	defer r.Body.Close()

	var body reqbody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 400, Error: err.Error(), Message: "Bad request body"})
		return
	}

	err = h.service.LoginWithEmail(body.Email)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 400, Error: err.Error(), Message: "Failed to send email"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Email sent"})
}
