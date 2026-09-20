package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"webshop/internal/repository"
)

type Mailer struct {
	settingsRepo *repository.SettingsRepo
	httpClient   *http.Client
}

func NewMailer(settingsRepo *repository.SettingsRepo) *Mailer {
	return &Mailer{
		settingsRepo: settingsRepo,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

type SMTPSettings struct {
	Host          string `json:"smtp_host"`
	Port          string `json:"smtp_port"`
	Username      string `json:"smtp_username"`
	Password      string `json:"smtp_password"`
	SenderName    string `json:"email_sender_name"`
	SenderAddress string `json:"email_sender_address"`
	ResendAPIKey  string `json:"resend_api_key"`
}

func (m *Mailer) getSettings(ctx context.Context) SMTPSettings {
	s := SMTPSettings{
		Port:          "587",
		SenderName:    "Aesthetic Studios",
		SenderAddress: "noreply@aesthetic.be",
	}

	allSettings, err := m.settingsRepo.GetAll(ctx)
	if err != nil {
		return s
	}

	if raw, ok := allSettings["smtp_host"]; ok { _ = json.Unmarshal(raw, &s.Host) }
	if raw, ok := allSettings["smtp_port"]; ok { _ = json.Unmarshal(raw, &s.Port) }
	if raw, ok := allSettings["smtp_username"]; ok { _ = json.Unmarshal(raw, &s.Username) }
	if raw, ok := allSettings["smtp_password"]; ok { _ = json.Unmarshal(raw, &s.Password) }
	if raw, ok := allSettings["email_sender_name"]; ok { _ = json.Unmarshal(raw, &s.SenderName) }
	if raw, ok := allSettings["email_sender_address"]; ok { _ = json.Unmarshal(raw, &s.SenderAddress) }
	if raw, ok := allSettings["resend_api_key"]; ok { _ = json.Unmarshal(raw, &s.ResendAPIKey) }

	return s
}

// SendHTML verstuurt een e-mail via SMTP, Resend of logt in simulatie-modus
func (m *Mailer) SendHTML(ctx context.Context, toEmail, subject, htmlBody string) error {
	s := m.getSettings(ctx)

	// 1. Resend API
	if s.ResendAPIKey != "" {
		payload := map[string]interface{}{
			"from":    fmt.Sprintf("%s <%s>", s.SenderName, s.SenderAddress),
			"to":      []string{toEmail},
			"subject": subject,
			"html":    htmlBody,
		}
		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(bodyBytes))
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+s.ResendAPIKey)
			req.Header.Set("Content-Type", "application/json")
			resp, err := m.httpClient.Do(req)
			if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				_ = resp.Body.Close()
				log.Printf("[RESEND E-MAIL] Succesvol verzonden naar %s (%s)", toEmail, subject)
				return nil
			}
		}
	}

	// 2. SMTP Server
	if s.Host != "" && s.Username != "" && s.Password != "" {
		fromHeader := fmt.Sprintf("%s <%s>", s.SenderName, s.SenderAddress)
		headers := make(map[string]string)
		headers["From"] = fromHeader
		headers["To"] = toEmail
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		message := ""
		for k, v := range headers {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\n" + htmlBody

		auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
		addr := net.JoinHostPort(s.Host, s.Port)

		// TLS configuratie
		tlsconfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err == nil {
			client, err := smtp.NewClient(conn, s.Host)
			if err == nil {
				if err = client.Auth(auth); err == nil {
					if err = client.Mail(s.SenderAddress); err == nil {
						if err = client.Rcpt(toEmail); err == nil {
							w, err := client.Data()
							if err == nil {
								_, _ = w.Write([]byte(message))
								_ = w.Close()
								_ = client.Quit()
								log.Printf("[SMTP E-MAIL] Succesvol verzonden via TLS naar %s", toEmail)
								return nil
							}
						}
					}
				}
			}
		}

		// Fallback gewone SMTP verbinding
		err = smtp.SendMail(addr, auth, s.SenderAddress, []string{toEmail}, []byte(message))
		if err == nil {
			log.Printf("[SMTP E-MAIL] Verzonden naar %s (%s)", toEmail, subject)
			return nil
		}
		log.Printf("[SMTP ERROR] Fout bij verzenden naar %s: %v", toEmail, err)
	}

	// 3. Veilige ontwikkelingssimulatie
	log.Printf("[E-MAIL SIMULATIE] Aan: %s | Onderwerp: %s", toEmail, subject)
	return nil
}

// SendOrderConfirmation stuurt een orderbevestiging
func (m *Mailer) SendOrderConfirmation(ctx context.Context, order *repository.Order, baseURL string) {
	recipient := "klant@shop.be"
	if order.GuestEmail != nil && *order.GuestEmail != "" {
		recipient = *order.GuestEmail
	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head><meta charset="utf-8"><style>body{font-family:-apple-system,BlinkMacSystemFont,sans-serif;color:#0f172a;line-height:1.6;background:#f8fafc;padding:20px;}.box{max-width:600px;margin:0 auto;background:#fff;border-radius:12px;padding:32px;border:1px solid #e2e8f0;}h1{font-size:24px;margin-bottom:8px;}.badge{display:inline-block;padding:4px 12px;border-radius:999px;background:#e2e8f0;font-size:12px;font-weight:700;}table{width:100%;border-collapse:collapse;margin:24px 0;}th,td{padding:12px;border-bottom:1px solid #f1f5f9;text-align:left;}th{color:#64748b;font-size:12px;text-transform:uppercase;}.total{font-size:18px;font-weight:800;}.btn{display:inline-block;padding:12px 24px;background:#0f172a;color:#fff;text-decoration:none;border-radius:8px;font-weight:600;margin-top:16px;}</style></head>
	<body>
	<div class="box">
		<span class="badge">Bestelbevestiging</span>
		<h1>Bedankt voor je aankoop!</h1>
		<p>We hebben bestelling <strong>{{.OrderNumber}}</strong> succesvol ontvangen en in behandeling genomen.</p>
		<table>
			<thead><tr><th>Artikel</th><th>Aantal</th><th>Prijs</th></tr></thead>
			<tbody>
				{{range .Items}}
				<tr><td>{{.ProductName}} ({{.VariantTitle}})</td><td>{{.Quantity}}</td><td>€ {{printf "%.2f" .TotalPrice}}</td></tr>
				{{end}}
			</tbody>
			<tfoot>
				<tr><td colspan="2">Verzendkosten</td><td>€ {{printf "%.2f" .ShippingCost}}</td></tr>
				<tr class="total"><td colspan="2">Totaalbedrag (incl. 21% Btw)</td><td>€ {{printf "%.2f" .TotalAmount}}</td></tr>
			</tfoot>
		</table>
		<p><a href="{{.BaseURL}}/track/{{.OrderNumber}}" class="btn">Volg Bestelling & Bekijk Factuur</a></p>
	</div>
	</body>
	</html>`

	t, _ := template.New("confirm").Parse(tmpl)
	buf := new(bytes.Buffer)
	_ = t.Execute(buf, map[string]interface{}{
		"OrderNumber":  order.OrderNumber,
		"Items":        order.Items,
		"ShippingCost": order.ShippingCost,
		"TotalAmount":  order.TotalAmount,
		"BaseURL":      strings.TrimSuffix(baseURL, "/"),
	})

	_ = m.SendHTML(ctx, recipient, fmt.Sprintf("Bestelbevestiging %s", order.OrderNumber), buf.String())
}

// SendShippingNotification stuurt bericht zodra het pakket onderweg is
func (m *Mailer) SendShippingNotification(ctx context.Context, order *repository.Order, trackingCode, carrier, baseURL string) {
	recipient := "klant@shop.be"
	if order.GuestEmail != nil && *order.GuestEmail != "" {
		recipient = *order.GuestEmail
	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head><meta charset="utf-8"><style>body{font-family:-apple-system,BlinkMacSystemFont,sans-serif;color:#0f172a;padding:20px;background:#f8fafc;}.box{max-width:600px;margin:0 auto;background:#fff;border-radius:12px;padding:32px;border:1px solid #e2e8f0;}h1{font-size:24px;}.tracking-box{background:#f1f5f9;padding:16px;border-radius:8px;margin:20px 0;}.btn{display:inline-block;padding:12px 24px;background:#0f172a;color:#fff;text-decoration:none;border-radius:8px;font-weight:600;}</style></head>
	<body>
	<div class="box">
		<h1>Je bestelling is onderweg! 📦</h1>
		<p>Goed nieuws! Bestelling <strong>{{.OrderNumber}}</strong> is overgedragen aan de koerier.</p>
		<div class="tracking-box">
			<strong>Vervoerder:</strong> {{.Carrier}}<br>
			<strong>Track & Trace code:</strong> <code>{{.TrackingCode}}</code>
		</div>
		<p><a href="{{.BaseURL}}/track/{{.OrderNumber}}" class="btn">Live Pakket Volgen</a></p>
	</div>
	</body>
	</html>`

	t, _ := template.New("shipping").Parse(tmpl)
	buf := new(bytes.Buffer)
	_ = t.Execute(buf, map[string]interface{}{
		"OrderNumber":  order.OrderNumber,
		"TrackingCode": trackingCode,
		"Carrier":      carrier,
		"BaseURL":      strings.TrimSuffix(baseURL, "/"),
	})

	_ = m.SendHTML(ctx, recipient, fmt.Sprintf("Je pakketje is onderweg! (%s)", order.OrderNumber), buf.String())
}