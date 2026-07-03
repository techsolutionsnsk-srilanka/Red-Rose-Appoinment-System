package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"redrose/backend/internal/domain"
)

// AppointmentRepository is the MySQL implementation of domain.AppointmentRepository.
type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

var _ domain.AppointmentRepository = (*AppointmentRepository)(nil)

func (r *AppointmentRepository) Create(ctx context.Context, a *domain.Appointment) error {
	const q = `
		INSERT INTO appointments
			(id, customer_name, customer_email, customer_phone, service,
			 starts_at, duration_min, status, notes, admin_notes, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		a.ID, a.CustomerName, a.CustomerEmail, a.CustomerPhone, a.Service,
		a.StartsAt.UTC(), a.DurationMin, a.Status, a.Notes, a.AdminNotes, a.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("create appointment: %w", err)
	}
	return nil
}

func (r *AppointmentRepository) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	const q = `
		SELECT id, customer_name, customer_email, customer_phone, service,
		       starts_at, duration_min, status, notes, admin_notes, created_by,
		       created_at, updated_at
		FROM appointments WHERE id = ?`
	row := r.db.QueryRowContext(ctx, q, id)
	a, err := scanAppointment(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get appointment: %w", err)
	}
	return a, nil
}

func (r *AppointmentRepository) List(ctx context.Context, f domain.AppointmentFilters) ([]domain.Appointment, error) {
	var (
		where []string
		args  []any
	)
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	if f.CustomerEmail != "" {
		where = append(where, "customer_email = ?")
		args = append(args, f.CustomerEmail)
	}
	if f.Service != "" {
		where = append(where, "service = ?")
		args = append(args, f.Service)
	}
	if !f.FromDate.IsZero() {
		where = append(where, "starts_at >= ?")
		args = append(args, f.FromDate.UTC())
	}
	if !f.ToDate.IsZero() {
		where = append(where, "starts_at <= ?")
		args = append(args, f.ToDate.UTC())
	}

	q := `SELECT id, customer_name, customer_email, customer_phone, service,
	             starts_at, duration_min, status, notes, admin_notes, created_by,
	             created_at, updated_at
	      FROM appointments`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY starts_at ASC"

	if f.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, f.Limit)
		if f.Offset > 0 {
			q += " OFFSET ?"
			args = append(args, f.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()

	var out []domain.Appointment
	for rows.Next() {
		a, err := scanAppointment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *AppointmentRepository) UpdateStatus(ctx context.Context, id string, status domain.AppointmentStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return requireAffected(res)
}

func (r *AppointmentRepository) UpdateAdminNotes(ctx context.Context, id, notes string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET admin_notes = ?, updated_at = ? WHERE id = ?`,
		notes, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update admin notes: %w", err)
	}
	return requireAffected(res)
}

func (r *AppointmentRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM appointments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete appointment: %w", err)
	}
	return requireAffected(res)
}

func (r *AppointmentRepository) GetStats(ctx context.Context) (*domain.AppointmentStats, error) {
	const q = `
		SELECT
			COUNT(*),
			COALESCE(SUM(status = 'pending'), 0),
			COALESCE(SUM(status = 'confirmed'), 0),
			COALESCE(SUM(status = 'completed'), 0),
			COALESCE(SUM(DATE(starts_at) = UTC_DATE() AND status IN ('pending','confirmed')), 0)
		FROM appointments`
	var s domain.AppointmentStats
	err := r.db.QueryRowContext(ctx, q).Scan(
		&s.TotalAppointments, &s.PendingAppointments, &s.ConfirmedAppointments,
		&s.CompletedAppointments, &s.UpcomingToday,
	)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &s, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanAppointment(s scanner) (*domain.Appointment, error) {
	var (
		a          domain.Appointment
		phone      sql.NullString
		notes      sql.NullString
		adminNotes sql.NullString
		createdBy  sql.NullString
	)
	err := s.Scan(
		&a.ID, &a.CustomerName, &a.CustomerEmail, &phone, &a.Service,
		&a.StartsAt, &a.DurationMin, &a.Status, &notes, &adminNotes, &createdBy,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.CustomerPhone = phone.String
	a.Notes = notes.String
	a.AdminNotes = adminNotes.String
	a.CreatedBy = createdBy.String
	return &a, nil
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("appointment not found")
	}
	return nil
}
