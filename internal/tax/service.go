package tax

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"webshop/internal/repository"
)

type Service struct {
	taxRepo      *repository.TaxRepo
	settingsRepo *repository.SettingsRepo
	httpClient   *http.Client
}

func NewService(taxRepo *repository.TaxRepo, settingsRepo *repository.SettingsRepo) *Service {
	return &Service{
		taxRepo:      taxRepo,
		settingsRepo: settingsRepo,
		httpClient:   &http.Client{Timeout: 6 * time.Second},
	}
}

type VIESResponse struct {
	IsValid   bool   `json:"isValid"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	UserError string `json:"userError,omitempty"`
}

type TaxCalculationResult struct {
	CountryCode     string  `json:"country_code"`
	TaxName         string  `json:"tax_name"`
	TaxRate         float64 `json:"tax_rate"`
	TaxAmount       float64 `json:"tax_amount"`
	SubtotalExclTax float64 `json:"subtotal_excl_tax"`
	SubtotalInclTax float64 `json:"subtotal_incl_tax"`
	IsReverseCharge bool    `json:"is_reverse_charge"`
	VatValid        bool    `json:"vat_valid"`
	VatCompanyName  string  `json:"vat_company_name,omitempty"`
}

func (s *Service) ValidateVIES(ctx context.Context, fullVatNumber string) (*VIESResponse, error) {
	clean := strings.ToUpper(regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(fullVatNumber, ""))
	if len(clean) < 4 {
		return &VIESResponse{IsValid: false}, nil
	}

	countryCode := clean[:2]
	vatNumber := clean[2:]

	url := fmt.Sprintf("https://ec.europa.eu/taxation_customs/vies/rest-api/ms/%s/vat/%s", countryCode, vatNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &VIESResponse{IsValid: false, UserError: "VIES service niet bereikbaar"}, nil
	}
	defer resp.Body.Close()

	var result VIESResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &VIESResponse{IsValid: false}, nil
	}

	return &result, nil
}

func (s *Service) CalculateTax(ctx context.Context, subtotal float64, countryCode, vatNumber string) (*TaxCalculationResult, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		countryCode = "BE"
	}

	rateInfo, err := s.taxRepo.GetRateForCountry(ctx, countryCode)
	if err != nil || rateInfo == nil {
		rateInfo = &repository.TaxRate{
			CountryCode: countryCode,
			Rate:        21.0,
			Name:        "Standaard Btw (21%)",
		}
	}

	result := &TaxCalculationResult{
		CountryCode: countryCode,
		TaxName:     rateInfo.Name,
		TaxRate:     rateInfo.Rate,
	}

	storeCountry := "BE"

	if vatNumber != "" && countryCode != storeCountry {
		vies, _ := s.ValidateVIES(ctx, vatNumber)
		if vies != nil && vies.IsValid {
			result.IsReverseCharge = true
			result.VatValid = true
			result.VatCompanyName = vies.Name
			result.TaxRate = 0.0
			result.TaxName = "Btw verlegd (Intracommunautaire levering art. 138 Richtlijn 2006/112/EG)"
		}
	}

	if result.TaxRate == 0.0 {
		result.TaxAmount = 0.0
		result.SubtotalExclTax = subtotal
		result.SubtotalInclTax = subtotal
	} else {
		subtotalExcl := subtotal / (1.0 + (result.TaxRate / 100.0))
		result.TaxAmount = subtotal - subtotalExcl
		result.SubtotalExclTax = subtotalExcl
		result.SubtotalInclTax = subtotal
	}

	return result, nil
}