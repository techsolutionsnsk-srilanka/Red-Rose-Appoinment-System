package email

import (
	"fmt"
	"html"
	"time"

	"redrose/backend/internal/domain"
)

// These renderers produce simple, self-contained HTML email bodies. Swap them
// for React Email / MJML templates when you outgrow inline HTML.

const brandColor = "#e11d48"

func layout(title, inner string) string {
	return fmt.Sprintf(`<!doctype html>
<html><body style="margin:0;background:#f6f6f7;font-family:system-ui,Segoe UI,Roboto,sans-serif;color:#1f2937">
  <div style="max-width:560px;margin:24px auto;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #eee">
    <div style="background:%s;color:#fff;padding:20px 24px;font-size:18px;font-weight:700">🌹 RedRose</div>
    <div style="padding:24px">
      <h2 style="margin:0 0 12px;font-size:20px">%s</h2>
      %s
    </div>
    <div style="padding:16px 24px;color:#9ca3af;font-size:12px;border-top:1px solid #f0f0f0">
      RedRose Appointments — this is an automated message.
    </div>
  </div>
</body></html>`, brandColor, html.EscapeString(title), inner)
}

func detailRows(a *domain.Appointment) string {
	when := a.StartsAt.Format(time.RFC1123)
	return fmt.Sprintf(`
	<table style="width:100%%;border-collapse:collapse;font-size:14px">
	  <tr><td style="padding:6px 0;color:#6b7280">Service</td><td style="padding:6px 0;text-align:right"><strong>%s</strong></td></tr>
	  <tr><td style="padding:6px 0;color:#6b7280">When</td><td style="padding:6px 0;text-align:right">%s</td></tr>
	  <tr><td style="padding:6px 0;color:#6b7280">Duration</td><td style="padding:6px 0;text-align:right">%d min</td></tr>
	  <tr><td style="padding:6px 0;color:#6b7280">Status</td><td style="padding:6px 0;text-align:right;text-transform:capitalize">%s</td></tr>
	</table>`,
		html.EscapeString(a.Service), html.EscapeString(when), a.DurationMin, html.EscapeString(string(a.Status)))
}

func renderCustomerConfirmation(a *domain.Appointment) string {
	inner := fmt.Sprintf(
		`<p>Hi %s,</p><p>Thanks for booking with RedRose. Your appointment has been received and is currently <strong>%s</strong>. We'll email you again once it is confirmed.</p>%s`,
		html.EscapeString(a.CustomerName), html.EscapeString(string(a.Status)), detailRows(a))
	return layout("Your appointment is booked", inner)
}

func renderStatusUpdate(a *domain.Appointment) string {
	inner := fmt.Sprintf(
		`<p>Hi %s,</p><p>The status of your appointment is now <strong style="text-transform:capitalize">%s</strong>.</p>%s`,
		html.EscapeString(a.CustomerName), html.EscapeString(string(a.Status)), detailRows(a))
	return layout("Appointment status updated", inner)
}

func renderAdminNewBooking(a *domain.Appointment, dashboardLink string) string {
	button := ""
	if dashboardLink != "" {
		button = fmt.Sprintf(
			`<p style="margin-top:16px"><a href="%s" style="background:%s;color:#fff;text-decoration:none;padding:10px 16px;border-radius:8px;display:inline-block">View in dashboard</a></p>`,
			html.EscapeString(dashboardLink), brandColor)
	}
	inner := fmt.Sprintf(
		`<p>A new appointment was booked by <strong>%s</strong> (%s).</p>%s%s`,
		html.EscapeString(a.CustomerName), html.EscapeString(a.CustomerEmail), detailRows(a), button)
	return layout("New booking received", inner)
}
