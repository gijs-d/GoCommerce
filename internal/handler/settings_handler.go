package handler

import (
	"encoding/json"
	"net/http"

	"webshop/internal/api"
	"webshop/internal/repository"
	"webshop/internal/theme"
)

type SettingsHandler struct {
	settingsRepo *repository.SettingsRepo
	themeManager *theme.ThemeManager
}

func NewSettingsHandler(settingsRepo *repository.SettingsRepo, themeManager *theme.ThemeManager) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: settingsRepo,
		themeManager: themeManager,
	}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsRepo.GetAll(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen instellingen")
		return
	}
	api.JSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige JSON data")
		return
	}

	for k, v := range payload {
		if err := h.settingsRepo.Set(r.Context(), k, v); err != nil {
			api.Error(w, http.StatusInternalServerError, "Fout bij opslaan van "+k)
			return
		}
	}

	api.Success(w, http.StatusOK, "Instellingen succesvol opgeslagen", nil)
}

func (h *SettingsHandler) ListThemes(w http.ResponseWriter, r *http.Request) {
	themes, err := h.themeManager.ListThemes(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij scannen van thema's: "+err.Error())
		return
	}
	api.JSON(w, http.StatusOK, themes)
}

func (h *SettingsHandler) ActivateTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ThemeID string `json:"theme_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ThemeID == "" {
		api.Error(w, http.StatusBadRequest, "Thema ID is verplicht")
		return
	}

	if err := h.themeManager.SetActiveThemeID(r.Context(), req.ThemeID); err != nil {
		api.Error(w, http.StatusInternalServerError, "Kon thema niet activeren")
		return
	}

	api.Success(w, http.StatusOK, "Thema succesvol geactiveerd: "+req.ThemeID, nil)
}