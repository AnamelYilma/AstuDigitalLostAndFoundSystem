# ASTU Digital Lost & Found

Modern web app for reporting, browsing, and approving lost/found items at ASTU.

## Stack
- Go 1.21+, Gin web framework
- GORM + PostgreSQL
- HTML templates with Tailwind (CDN) + light custom CSS
- Gorilla sessions (cookie store), bcrypt password hashing
- CSRF middleware

## Folder Structure (key parts)
- `main.go` – bootstrap, routes, template loading, middleware
- `internal/handler` – HTTP handlers (auth, item, admin)
- `internal/service` – business rules (approvals, claims, notifications)
- `internal/repository` – DB access via GORM
- `internal/model` – User, Item, ItemImage, Claim, Notification structs
- `internal/middleware` – auth/session, admin guard, CSRF
- `templates` – HTML (layout, index, dashboard, report form, report view/cards, item detail, admin pages)
- `static/css/style.css` – minor overrides (Tailwind loaded via CDN)
- `static/uploads` – saved item images

## Run Locally
1) Set env (or `.env`):
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpass
DB_NAME=lostfound
DB_SSLMODE=disable
# optional: DATABASE_URL=postgres://user:pass@host:port/db?sslmode=disable
SESSION_SECRET=change_me_32_chars
```
2) Start Postgres.
3) Run server:
```
go run .
```
Logs: "Server starting on http://localhost:8080".

## Default Accounts
- Admin auto-created if none: `admin / admin123`
- Students register with name, student ID, phone, password; login by student ID + password.

## Core Features & Flow
- Report Lost/Found: `/report/new?type=lost|found` (auth). Found requires at least one photo.
- Report View (public browse): `/report` shows all approved posts with filters (type, category, location, color, since-date).
- Item Detail: hides reporter contact/location for found posts until admin-approved request.
- Claims/Requests: users submit claim or found-match; admin approves/rejects; notifications send contact info to both sides on approval.
- Admin: approves/rejects posts, manages claims, removes items, sees stats.

## Security
- Password hashing (bcrypt in `pkg/utils/hash.go`).
- Session cookies (HttpOnly, SameSite=Lax, optional Secure via env).
- CSRF tokens on all POST forms (`internal/middleware/csrf_middleware.go`).
- Role-based access: `/admin/*` guarded by admin middleware; report form, claims, notifications require login.
- Basic input validation: item type/category/location/color enums, image size/type checks.

## Data Model (high level)
- User: name, student_id (unique), phone, email, role, password hash.
- Item: type(lost/found), title, category, color, brand, location, date, description, approval_status, status(open/claimed), images.
- Claim: request_type (claim_request/found_match_request), status, admin remarks.
- Notification: user_id, title, message, is_read.

## Routes (human-facing)
- Public: `GET /` home, `GET /report` list, `GET /item/:id` detail
- Auth: `GET/POST /login`, `GET/POST /register`, `GET /logout`
- Protected: `GET /report/new`, `POST /report/new`, `POST /claim`, `GET/POST /notifications`, `GET /dashboard`
- Admin: `GET /admin/dashboard`, `GET/POST /admin/items`, `GET/POST /admin/claims`

## UI Notes
- Tailwind via CDN; templates use modern cards, compact filters, and drag-and-drop styled upload on the report form.
- `static/css/style.css` only tweaks small defaults; width is full-page (no fixed max container in layout).

## How It Works (example)
1) Student reports found item ? item saved as pending ? admin approves ? item appears in Report View.
2) Owner submits claim ? admin approves ? both parties get notification with contact info ? they connect offline.
3) Admin can reject posts/claims with remarks; posters/claimants get notified.

## Troubleshooting
- If Tailwind styles don’t show, hard-refresh (Ctrl+F5); templates load CDN script in `layout.html`.
- If DB fails: verify env vars and Postgres is running; check connection string.
- If build errors mention `search.html`: ensure `main.go` does not load removed templates (already adjusted).

## Development Tips
- Templates are server-rendered; update `templates/*.html` and restart `go run .`.
- Images are stored under `static/uploads`; ensure write perms.
- Use `go build ./...` to verify after edits.
