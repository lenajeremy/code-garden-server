package handlers

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/auth"
	"code-garden-server/utils"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
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
	Email      string `json:"email"`
	ClientHost string `json:"clientHost"`
}

func (h *AuthHandler) VerifyUserEmail(w http.ResponseWriter, r *http.Request) {
	tokenString := r.PathValue("token")

	if tokenString == "" {
		utils.WriteRes(w, utils.Response{Status: 400, Message: "Bad request: Invalid token", Error: "Invalid token"})
		return
	}

	_, err := h.service.VerifyUserEmail(tokenString)

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
	if err != nil || body.Email == "" {
		utils.WriteRes(w, utils.Response{Status: 400, Error: "Invalid email", Message: "Bad request body"})
		return
	}

	err = h.service.RegisterWithEmail(body.Email, body.ClientHost)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 500, Error: err.Error(), Message: "Failed to create account"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Success! Please check your mail to verify your account"})
}

func (h *AuthHandler) RegisterWithPassword(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	type requestBody struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		ClientHost string `json:"clientHost"`
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Password == "" || body.Email == "" {
		utils.WriteRes(w, utils.Response{Error: "Invalid password/email", Status: 400, Message: "Bad request"})
		return
	}

	err = h.service.RegisterWithPassword(body.Email, body.Password, body.ClientHost)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 500, Error: err.Error(), Message: "Failed to sign user in"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Account created successfully!"})
}

func (h *AuthHandler) LoginWithEmail(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Email == "" {
		utils.WriteRes(w, utils.Response{Status: 400, Error: "Invalid email", Message: "Bad request body"})
		return
	}

	err = h.service.LoginWithEmail(body.Email, body.ClientHost)
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

func (h *AuthHandler) SignInWithToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	if token == "" {
		utils.WriteRes(w, utils.Response{Status: 400, Error: "invalid token", Message: "Bad request. Invalid token"})
		return
	}

	jwtToken, err := h.service.GenerateJwtFromToken(token)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 500, Error: err.Error(), Message: "Failed to sign user in"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Successfully signed in!", Data: map[string]string{"token": jwtToken}})
}

func (h *AuthHandler) LoginWithPassword(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Error: err.Error(), Status: 400, Message: "Bad request"})
		return
	}

	if body.Password == "" || body.Email == "" {
		utils.WriteRes(w, utils.Response{Error: "Invalid password or email address", Status: 400, Message: "Bad request"})
		return
	}

	jwtToken, err := h.service.LoginWithPassword(body.Email, body.Password)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: 500, Error: err.Error(), Message: "Failed to sign user in"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: 200, Message: "Successfully signed in!", Data: map[string]string{"token": jwtToken}})
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	type emailReq struct {
		Email string `json:"email"`
		Host  string `json:"host"`
	}

	var body emailReq
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Email == "" || body.Host == "" {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: "Bad request", Message: "Empty email/host"})
		return
	}

	err = h.service.SendResetPasswordEmail(body.Email, body.Host)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Error: err.Error(), Message: "Internal server error"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Data: "Done!", Message: "Success! Check your mail"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	type inputRes struct {
		NewPassword        string `json:"newPassword"`
		ConfirmNewPassword string `json:"confirmNewPassword"`
		ValidationToken    string `json:"validationToken"`
	}

	var body inputRes
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: err.Error(), Message: "Bad request"})
		return
	}

	if body.NewPassword == "" ||
		body.ConfirmNewPassword == "" ||
		body.ValidationToken == "" {
		if body.ValidationToken == "" {
			utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: "validation token missing", Message: "Bad request"})
		} else {
			utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: "password/confirmation missing", Message: "Bad request"})
		}
		return
	}

	if body.NewPassword != body.ConfirmNewPassword {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: "password and confirmation password mismatch", Message: "Bad request"})
		return
	}

	err = h.service.ResetUserPassword(body.ValidationToken, body.NewPassword)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Message: "Failed to reset password", Error: err.Error()})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Data: "Done", Message: "Password reset successfully. Proceed to login"})
}
