package handler

import (
	"encoding/json"
	"net/http"

	"webshop/internal/api"
	"webshop/internal/email"
)

type EmailHandler struct {
	mailer *email.Mailer
}

func NewEmailHandler(mailer *email.Mailer) *EmailHandler {
	return &EmailHandler{mailer: mailer}
}

func (h *EmailHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		api.Error(w, http.StatusBadRequest, "E-mailadres is verplicht")
		return
	}

	subject := "Test e-mail vanuit je WooCommerce Go webshop"
	body := `
		<h2>Verbinding Geslaagd! 🎉</h2>
		<p>Je e-mailserver (SMTP / Resend) is succesvol geconfigureerd en klaar om bestelbevestigingen te versturen.</p>
	`
	err := h.mailer.SendHTML(r.Context(), req.Email, subject, body)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Verzenden mislukt: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Test e-mail succesvol verzonden naar "+req.Email, nil)
}