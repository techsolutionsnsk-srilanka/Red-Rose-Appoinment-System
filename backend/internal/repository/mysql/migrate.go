package mysql

import (
	"context"
	"database/sql"
	"fmt"
)

// EnsureSchema creates the tables the application needs if they do not exist.
// It is safe to call on every startup. For richer, versioned migrations use the
// .sql files in backend/migrations with your migration tool of choice.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS appointments (
	id             VARCHAR(64)  NOT NULL PRIMARY KEY,
	customer_name  VARCHAR(255) NOT NULL,
	customer_email VARCHAR(255) NOT NULL,
	customer_phone VARCHAR(64)      NULL,
	service        VARCHAR(255) NOT NULL,
	starts_at      DATETIME     NOT NULL,
	duration_min   INT          NOT NULL DEFAULT 30,
	status         VARCHAR(32)  NOT NULL DEFAULT 'pending',
	notes          TEXT             NULL,
	admin_notes    TEXT             NULL,
	created_by     VARCHAR(255)     NULL,
	created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	INDEX idx_appointments_starts_at (starts_at),
	INDEX idx_appointments_status (status),
	INDEX idx_appointments_email (customer_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("ensure appointments schema: %w", err)
	}
	return nil
}
