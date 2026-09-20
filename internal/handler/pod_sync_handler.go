package handler

import (
	"encoding/json"
	"net/http"

	"webshop/internal/api"
	"webshop/internal/pod"
	"webshop/internal/repository"
)

type PODSyncHandler struct {
	gelatoClient *pod.GelatoClient
	shippingRepo *repository.ShippingRepo
}

func NewPODSyncHandler(gelatoClient *pod.GelatoClient, shippingRepo *repository.ShippingRepo) *PODSyncHandler {
	return &PODSyncHandler{
		gelatoClient: gelatoClient,
		shippingRepo: shippingRepo,
	}
}

type LiveQuotesRequest struct {
	Destination pod.PODAddress `json:"destination"`
	Items       []pod.PODItem  `json:"items"`
	Subtotal    float64        `json:"subtotal"`
}

// GetLiveShipmentQuotes combineert vaste winkel-verzendopties met realtime POD-offertes
func (h *PODSyncHandler) GetLiveShipmentQuotes(w http.ResponseWriter, r *http.Request) {
	var req LiveQuotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "Ongeldige aanvraag voor verzendtarieven")
		return
	}

	var allQuotes []pod.ShipmentQuote

	// 1. Haal de geconfigureerde winkel verzendmethoden op uit de database
	shopMethods, _ := h.shippingRepo.ListActive(r.Context())
	for _, m := range shopMethods {
		cost := m.Cost
		if m.FreeThreshold != nil && req.Subtotal >= *m.FreeThreshold {
			cost = 0.00
		}
		allQuotes = append(allQuotes, pod.ShipmentQuote{
			ID:          m.ID,
			Title:       m.Title,
			Description: *m.Description,
			Cost:        cost,
			Currency:    "EUR",
			MinDays:     1,
			MaxDays:     3,
			Provider:    "shop",
		})
	}

	// 2. Haal realtime offertes op van Gelato indien er POD artikelen in het mandje zitten
	var podItems []pod.PODItem
	for _, it := range req.Items {
		if it.PODVariantID != "" {
			podItems = append(podItems, it)
		}
	}

	if len(podItems) > 0 {
		gelatoQuotes, err := h.gelatoClient.GetShipmentQuotes(r.Context(), req.Destination, podItems)
		if err == nil && len(gelatoQuotes) > 0 {
			allQuotes = append(allQuotes, gelatoQuotes...)
		}
	}

	api.JSON(w, http.StatusOK, allQuotes)
}

// SyncCatalog triggert de import van Gelato POD artikelen
func (h *PODSyncHandler) SyncCatalog(w http.ResponseWriter, r *http.Request) {
	count, err := h.gelatoClient.SyncCatalog(r.Context())
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "Fout bij synchroniseren van catalogus: "+err.Error())
		return
	}

	api.Success(w, http.StatusOK, "Gelato catalogus succesvol gesynchroniseerd", map[string]int{
		"synced_products": count,
	})
}