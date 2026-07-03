package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"redrose/backend/internal/domain"
)

const resendEndpoint = "https://api.resend.com/emails"

// ResendNotifier sends appointment notifications via the Resend email API.
// It implements domain.AppointmentNotifier.
type ResendNotifier struct {
	apiKey     string
	fromEmail  string
	adminEmail string
	adminURL   string
	httpClient *http.Client
}

// NewResendNotifier creates a Resend-backed notifier.
//   - apiKey:     Resend API key (RESEND_API_KEY)
//   - fromEmail:  verified sender, e.g. "RedRose <noreply@example.com>" (FROM_EMAIL)
//   - adminEmail: internal recipient for new-booking alerts (ADMIN_EMAIL)
//   - adminURL:   base URL of the admin dashboard, for "view in dashboard" links (ADMIN_URL)
func NewResendNotifier(apiKey, fromEmail, adminEmail, adminURL string) *ResendNotifier {
	return &ResendNotifier{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		adminEmail: adminEmail,
		adminURL:   adminURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

var _ domain.AppointmentNotifier = (*ResendNotifier)(nil)

// SendBookingConfirmation emails the customer their confirmation and, if an
// admin address is configured, alerts the admin about the new booking.
func (n *ResendNotifier) SendBookingConfirmation(ctx context.Context, a *domain.Appointment) error {
	if a == nil {
		return fmt.Errorf("appointment is nil")
	}

	if a.CustomerEmail != "" {
		subject := fmt.Sprintf("Your RedRose appointment for %s is booked", a.Service)
		if err := n.send(ctx, a.CustomerEmail, subject, renderCustomerConfirmation(a)); err != nil {
			return fmt.Errorf("send customer confirmation: %w", err)
		}
	}

	if n.adminEmail != "" {
		link := ""
		if n.adminURL != "" {
			link = strings.TrimRight(n.adminURL, "/") + "/appointments/" + a.ID
		}
		subject := fmt.Sprintf("New booking: %s — %s", a.Service, a.StartsAt.Format("Mon 2 Jan 15:04"))
		if err := n.send(ctx, n.adminEmail, subject, renderAdminNewBooking(a, link)); err != nil {
			return fmt.Errorf("send admin alert: %w", err)
		}
	}

	return nil
}

// SendStatusUpdate emails the customer that their appointment status changed.
func (n *ResendNotifier) SendStatusUpdate(ctx context.Context, a *domain.Appointment) error {
	if a == nil || a.CustomerEmail == "" {
		return nil
	}
	subject := fmt.Sprintf("Your RedRose appointment is now %s", a.Status)
	return n.send(ctx, a.CustomerEmail, subject, renderStatusUpdate(a))
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (n *ResendNotifier) send(ctx context.Context, to, subject, html string) error {
	if n.apiKey == "" {
		return fmt.Errorf("resend api key is not configured")
	}

	payload, err := json.Marshal(resendRequest{
		From:    n.fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+n.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
