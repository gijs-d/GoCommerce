package handler

import (
	"net/http"

	"webshop/internal/api"
	"webshop/internal/repository"
)

type WidgetHandler struct {
	widgetRepo *repository.WidgetRepo
}

func NewWidgetHandler(widgetRepo *repository.WidgetRepo) *WidgetHandler {
	return &WidgetHandler{widgetRepo: widgetRepo}
}

func (h *WidgetHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.widgetRepo.ListActive(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij inladen van homepage widgets")
		return
	}

	api.JSON(w, http.StatusOK, widgets)
}