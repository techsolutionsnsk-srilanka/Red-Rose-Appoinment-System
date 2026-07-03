-- 001_init.sql — RedRose appointments schema
-- Applied automatically at startup by mysql.EnsureSchema; kept here as the
-- source of truth for versioned migrations.

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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
