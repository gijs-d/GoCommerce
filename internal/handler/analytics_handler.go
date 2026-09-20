package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"webshop/internal/api"
	"webshop/internal/repository"
)

type AnalyticsHandler struct {
	analyticsRepo *repository.AnalyticsRepo
}

func NewAnalyticsHandler(analyticsRepo *repository.AnalyticsRepo) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsRepo: analyticsRepo}
}

func (h *AnalyticsHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && d > 0 {
		days = d
	}

	data, err := h.analyticsRepo.GetOverview(r.Context(), days)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij ophalen analytics: "+err.Error())
		return
	}

	api.JSON(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	csvBytes, err := h.analyticsRepo.GenerateCSVExport(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij genereren CSV export")
		return
	}

	fileName := fmt.Sprintf("verkooprapport-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}