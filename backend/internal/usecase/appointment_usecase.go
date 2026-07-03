package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"redrose/backend/internal/domain"

	"github.com/google/uuid"
)

// AppointmentUseCase holds the business logic for appointments. It depends only
// on domain interfaces, so the storage and email layers can be swapped freely.
type AppointmentUseCase struct {
	repo     domain.AppointmentRepository
	notifier domain.AppointmentNotifier
}

func NewAppointmentUseCase(repo domain.AppointmentRepository, notifier domain.AppointmentNotifier) *AppointmentUseCase {
	return &AppointmentUseCase{repo: repo, notifier: notifier}
}

func (uc *AppointmentUseCase) CreateAppointment(ctx context.Context, a *domain.Appointment) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	if a.Status == "" {
		a.Status = domain.AppointmentStatusPending
	}
	if a.DurationMin == 0 {
		a.DurationMin = 30
	}

	if err := uc.validate(a); err != nil {
		return err
	}

	if err := uc.repo.Create(ctx, a); err != nil {
		return err
	}

	// Best-effort confirmation email; never blocks or fails the booking.
	if uc.notifier != nil {
		go func(appt domain.Appointment) {
			bg, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := uc.notifier.SendBookingConfirmation(bg, &appt); err != nil {
				log.Printf("⚠️  booking confirmation for %s failed: %v", appt.ID, err)
			} else {
				log.Printf("📧 booking confirmation sent for %s", appt.ID)
			}
		}(*a)
	}

	return nil
}

func (uc *AppointmentUseCase) GetAppointment(ctx context.Context, id string) (*domain.Appointment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *AppointmentUseCase) ListAppointments(ctx context.Context, filters domain.AppointmentFilters) ([]domain.Appointment, error) {
	return uc.repo.List(ctx, filters)
}

func (uc *AppointmentUseCase) UpdateStatus(ctx context.Context, id string, status domain.AppointmentStatus) error {
	if !status.Valid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	if err := uc.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if uc.notifier != nil {
		go func() {
			bg, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			appt, err := uc.repo.GetByID(bg, id)
			if err != nil || appt == nil {
				log.Printf("⚠️  status email: could not load appointment %s: %v", id, err)
				return
			}
			if err := uc.notifier.SendStatusUpdate(bg, appt); err != nil {
				log.Printf("⚠️  status email for %s failed: %v", id, err)
			} else {
				log.Printf("📧 status email sent for %s (%s)", id, status)
			}
		}()
	}

	return nil
}

func (uc *AppointmentUseCase) UpdateAdminNotes(ctx context.Context, id, notes string) error {
	return uc.repo.UpdateAdminNotes(ctx, id, notes)
}

func (uc *AppointmentUseCase) DeleteAppointment(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *AppointmentUseCase) GetStats(ctx context.Context) (*domain.AppointmentStats, error) {
	return uc.repo.GetStats(ctx)
}

func (uc *AppointmentUseCase) validate(a *domain.Appointment) error {
	if strings.TrimSpace(a.CustomerName) == "" {
		return fmt.Errorf("customer name is required")
	}
	if strings.TrimSpace(a.CustomerEmail) == "" {
		return fmt.Errorf("customer email is required")
	}
	if strings.TrimSpace(a.Service) == "" {
		return fmt.Errorf("service is required")
	}
	if a.StartsAt.IsZero() {
		return fmt.Errorf("starts_at is required")
	}
	if a.StartsAt.Before(time.Now().Add(-time.Minute)) {
		return fmt.Errorf("starts_at must be in the future")
	}
	if a.DurationMin <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	return nil
}
