package payment

type PaymentMethodInfo struct {
	ID          string `json:"id"`          // bijv. 'mollie_ideal', 'mollie_bancontact', 'stripe_card', 'mock'
	Title       string `json:"title"`       // 'iDEAL', 'Bancontact', 'Creditcard', 'Directe Testbetaling'
	Description string `json:"description"` // Toelichting voor klant
	Provider    string `json:"provider"`    // 'mollie', 'stripe', 'mock'
	Icon        string `json:"icon"`
}

type PaymentInitResult struct {
	Success      bool   `json:"success"`
	RedirectURL  string `json:"redirect_url,omitempty"` // URL naar Mollie/Stripe checkout
	PaymentID    string `json:"payment_id,omitempty"`
	RequiresWait bool   `json:"requires_wait"`
	OrderNumber  string `json:"order_number"`
}