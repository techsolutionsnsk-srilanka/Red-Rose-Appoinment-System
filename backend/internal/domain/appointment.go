package domain

import (
	"context"
	"time"
)

// AppointmentStatus represents the lifecycle state of an appointment.
type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
)

// Appointment represents a customer booking in the RedRose system.
type Appointment struct {
	ID            string            `json:"id"`
	CustomerName  string            `json:"customer_name"`
	CustomerEmail string            `json:"customer_email"`
	CustomerPhone string            `json:"customer_phone"`
	Service       string            `json:"service"`
	StartsAt      time.Time         `json:"starts_at"`
	DurationMin   int               `json:"duration_min"`
	Status        AppointmentStatus `json:"status"`
	Notes         string            `json:"notes"`
	AdminNotes    string            `json:"admin_notes"`
	CreatedBy     string            `json:"created_by"` // Clerk user ID, when booked from the admin
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// AppointmentNotifier sends appointment-related notifications (e.g. email).
// Implemented in the infrastructure layer; must never block critical flow.
type AppointmentNotifier interface {
	// SendBookingConfirmation notifies the customer (and admin) of a new booking.
	SendBookingConfirmation(ctx context.Context, a *Appointment) error
	// SendStatusUpdate notifies the customer that their appointment status changed.
	SendStatusUpdate(ctx context.Context, a *Appointment) error
}

// AppointmentRepository defines persistence for appointments.
type AppointmentRepository interface {
	Create(ctx context.Context, a *Appointment) error
	GetByID(ctx context.Context, id string) (*Appointment, error)
	List(ctx context.Context, filters AppointmentFilters) ([]Appointment, error)
	UpdateStatus(ctx context.Context, id string, status AppointmentStatus) error
	UpdateAdminNotes(ctx context.Context, id string, notes string) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*AppointmentStats, error)
}

// AppointmentFilters for querying appointments.
type AppointmentFilters struct {
	Status        AppointmentStatus
	CustomerEmail string
	Service       string
	FromDate      time.Time
	ToDate        time.Time
	Limit         int
	Offset        int
}

// AppointmentStats for the admin dashboard.
type AppointmentStats struct {
	TotalAppointments     int `json:"total_appointments"`
	PendingAppointments   int `json:"pending_appointments"`
	ConfirmedAppointments int `json:"confirmed_appointments"`
	CompletedAppointments int `json:"completed_appointments"`
	UpcomingToday         int `json:"upcoming_today"`
}

// Valid reports whether s is a recognised status.
func (s AppointmentStatus) Valid() bool {
	switch s {
	case AppointmentStatusPending, AppointmentStatusConfirmed,
		AppointmentStatusCancelled, AppointmentStatusCompleted:
		return true
	default:
		return false
	}
}
