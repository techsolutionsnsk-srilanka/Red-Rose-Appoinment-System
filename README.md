# 🌹 RedRose Appointment System

Appointment booking platform built in the **v0-numerore architecture**: three
self-contained apps side by side, a Go backend in Clean Architecture, Clerk auth,
Resend email, MySQL, and Docker/Railway deploy configs per app.

```
.
├── admin/      Next.js 15 admin dashboard — Clerk auth, Tailwind v4, shadcn-style UI (port 3001)
├── web/        Next.js 15 public booking site — no auth, books via the API   (port 3000)
└── backend/    Go + Gin REST API — Clean Architecture, MySQL, Clerk JWT verify, Resend email (port 8080)
```

## Architecture

### backend/ — Clean Architecture (domain → usecase → repository → handler)

```
backend/
├── cmd/api/main.go                 Composition root: wires DB, repos, usecases, handlers
├── internal/
│   ├── domain/                     Entities + Repository/Notifier interfaces (no deps)
│   │   └── appointment.go
│   ├── usecase/                    Business logic, depends only on domain interfaces
│   │   └── appointment_usecase.go
│   ├── repository/mysql/           MySQL implementation of domain.AppointmentRepository
│   │   ├── appointment_repository.go
│   │   └── migrate.go              EnsureSchema() runs on startup
│   ├── handler/http/               Gin handlers (transport layer)
│   │   └── appointment_handler.go
│   ├── middleware/auth.go          Verifies Clerk RS256 JWTs against the issuer JWKS
│   └── email/                      Resend notifier + HTML templates
│       ├── resend.go
│       └── templates.go
├── migrations/001_init.sql
├── Dockerfile · railway.json · .env.example
```

The **dependency rule** points inward: `handler → usecase → domain ← repository/mysql`.
Swap MySQL for Postgres by writing a new `domain.AppointmentRepository` — nothing
else changes.

### admin/ — Next.js (Clerk)

- `middleware.ts` — Clerk route protection (everything except `/login`, `/signup`).
- `lib/auth.ts` — pulls the Clerk session token; `lib/api.ts` sends it as a
  `Bearer` token to the Go API.
- `app/(auth)/` — Clerk `<SignIn>` / `<SignUp>` pages.
- `app/(dashboard)/` — stats dashboard + appointment management table.

### web/ — Next.js (public)

Landing page + `/book` form that POSTs to the **public** `POST /api/appointments`
endpoint. Triggers the Resend confirmation email.

## Auth flow

1. Admin signs in with Clerk in `admin/`.
2. `lib/api.ts` attaches the Clerk session JWT as `Authorization: Bearer <token>`.
3. Backend `AuthMiddleware` fetches the Clerk instance's JWKS from
   `{iss}/.well-known/jwks.json`, verifies the RS256 signature, and exposes the
   Clerk `user_id`/`email` to handlers.
4. Public booking (`web/`) hits the unauthenticated `POST /api/appointments`.

## API

| Method | Path                              | Auth   | Description                 |
| ------ | --------------------------------- | ------ | --------------------------- |
| GET    | `/health`                         | public | Health check                |
| POST   | `/api/appointments`               | public | Book an appointment + email |
| GET    | `/api/appointments`               | admin  | List (filterable)           |
| GET    | `/api/appointments/stats`         | admin  | Dashboard stats             |
| GET    | `/api/appointments/:id`           | admin  | Get one                     |
| PUT    | `/api/appointments/:id/status`    | admin  | Update status + email       |
| PUT    | `/api/appointments/:id/notes`     | admin  | Update admin notes          |
| DELETE | `/api/appointments/:id`           | admin  | Delete                      |

## Running locally

### 0. Prerequisites
Node 20+, Go 1.25+, a MySQL instance, a [Clerk](https://clerk.com) app, and
(optionally) a [Resend](https://resend.com) key.

### 1. Backend
```bash
cd backend
cp .env.example .env         # set DB_* creds; set AUTH_DISABLED=true for quick local dev
go mod tidy
go run ./cmd/api             # :8080  (auto-creates the appointments table)
```

### 2. Admin
```bash
cd admin
cp .env.example .env.local   # set Clerk keys + NEXT_PUBLIC_API_URL
npm install
npm run dev                  # http://localhost:3001
```

### 3. Web
```bash
cd web
cp .env.example .env.local   # set NEXT_PUBLIC_API_URL
npm install
npm run dev                  # http://localhost:3000
```

## Deployment (Railway)

Each app ships a `Dockerfile` and `railway.json` — deploy `admin/`, `web/`, and
`backend/` as three Railway services. The Next.js apps use `output: 'standalone'`.

## Extending the template

- Add a domain (e.g. `service`, `customer`): create `domain/<x>.go` with its
  interfaces, a `usecase`, a `repository/mysql` impl, and a `handler/http` — then
  wire them in `cmd/api/main.go`. The appointment slice is the reference pattern.
- Gate admin-only actions with Clerk roles/permissions inside `AuthMiddleware`.
- Replace inline email HTML in `email/templates.go` with React Email / MJML.
