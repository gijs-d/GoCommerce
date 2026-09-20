package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"webshop/internal/api"
	"webshop/internal/auth"
	"webshop/internal/config"
	"webshop/internal/middleware"
	"webshop/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepo
	cfg      *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepo, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, cfg: cfg}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	isSecure := r.TLS != nil || h.cfg.AppEnv == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige aanvraag")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" || len(req.Password) < 6 || req.FullName == "" {
		api.Error(w, http.StatusBadRequest, "Vul alle verplichte velden in (wachtwoord minstens 6 tekens)")
		return
	}

	existing, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Serverfout bij controleren e-mailadres")
		return
	}
	if existing != nil {
		api.Error(w, http.StatusConflict, "Dit e-mailadres is al in gebruik")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij verwerken van wachtwoord")
		return
	}

	user, err := h.userRepo.CreateUser(r.Context(), req.Email, hash, req.FullName, "customer")
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon gebruiker niet aanmaken")
		return
	}

	token, err := h.userRepo.CreateSession(r.Context(), user.ID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon sessie niet aanmaken")
		return
	}

	h.setSessionCookie(w, r, token, int(auth.SessionDuration().Seconds()))
	api.Success(w, http.StatusCreated, "Registratie geslaagd", user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige aanvraag")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		api.Error(w, http.StatusUnauthorized, "Onjuist e-mailadres of wachtwoord")
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		api.Error(w, http.StatusUnauthorized, "Onjuist e-mailadres of wachtwoord")
		return
	}

	token, err := h.userRepo.CreateSession(r.Context(), user.ID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon sessie niet starten")
		return
	}

	h.setSessionCookie(w, r, token, int(auth.SessionDuration().Seconds()))
	api.Success(w, http.StatusOK, "Succesvol ingelogd", map[string]interface{}{
		"user": map[string]string{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
		},
		"token": token,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetSessionTokenFromContext(r.Context())
	if token == "" {
		token = middleware.ExtractSessionToken(r)
	}

	if token != "" {
		_ = h.userRepo.DeleteSession(r.Context(), token)
	}

	h.setSessionCookie(w, r, "", -1)
	api.Success(w, http.StatusOK, "Succesvol uitgelogd", nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	currentUser := middleware.GetUserFromContext(r.Context())
	if currentUser == nil {
		api.Error(w, http.StatusUnauthorized, "Niet ingelogd")
		return
	}

	api.JSON(w, http.StatusOK, currentUser)
}