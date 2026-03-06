# ASTU Digital Lost & Found

A campus-ready web app for reporting, browsing, and matching lost/found items at Adama Science & Technology University (ASTU). Students submit posts, admins approve them, and both sides get notified when a match is approved.

---

## Why it exists
- Remove paper logs and scattered chats; give students one place to report.
- Keep sensitive contact info hidden until an admin approves a match.
- Make admins faster: pending queues, filters, and one-click approvals.

## How it works (flow)
1) **Report** – Logged-in student opens `/report/new?type=lost|found`, enters details, uploads up to 5 photos (found posts require at least one). Post is stored as `pending`.
2) **Review** – Admin dashboard lists pending posts. Approve ➜ post becomes public; Reject ➜ poster sees remark.
3) **Browse** – Public `/report` shows only approved posts with filters (type, category, location, color, date).
4) **Request** – On an item page, a different user files a claim (or “found match” if the original post is lost). Multiple active claims per user/post are prevented.
5) **Decision** – Admin approves a request ➜ item is marked `claimed` and both parties receive notifications with each other’s contact info. Rejection sends the requester a remark.
6) **Notify** – Users see unread counts in the nav; `/notifications` lets them read/mark all.

---

## Tech stack
- **Backend:** Go 1.21+, Gin, GORM (PostgreSQL)
- **Views:** Server-rendered HTML templates (Tailwind via CDN with local CSS tweaks)
- **Sessions & Auth:** Gorilla sessions (cookie store), bcrypt password hashing
- **Security middleware:** CSRF tokens, role-based guards (student/admin)

---

## Security highlights
- Passwords hashed with bcrypt.
- CSRF protection on all state-changing requests (including logout as POST).
- Session cookie: HttpOnly + SameSite=Lax, `Secure` when `COOKIE_SECURE=true`.
- DSN masking in logs; production lowers GORM logging to WARN to avoid PII in SQL traces.
- Upload hardening: MIME sniffing (JPEG/PNG), 5 MB per file, max 5 images, orphan file cleanup on delete.
- Public listing shows only approved posts; contact/location on found items stay hidden until an approved match.

---

## Configuration
Create `.env` (or set env vars):
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpass
DB_NAME=lostfound
DB_SSLMODE=disable
# or DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable

SESSION_SECRET=change_me_32_chars
GO_ENV=development         # set to production in prod
COOKIE_SECURE=false        # true when behind HTTPS
```

---

## Running locally
```bash
go mod download
go run .
# server listens on :8080 by default (PORT env overrides)
```
The app auto-migrates tables and, if no admin exists, seeds `admin / admin123` (change immediately in prod).

---

## Using the app
- **Register/Login:** `/register`, `/login` (login via student ID + password).
- **Report item:** `/report/new?type=lost` or `found` (auth required).
- **Browse:** `/report` with filters; click a card to see details.
- **Request/claim:** On an item page, submit the claim form (cannot claim your own post).
- **Notifications:** `/notifications` to read/mark all.
- **Admin:** `/admin/dashboard`, `/admin/items`, `/admin/claims` (admins only).

---

## Data model (simplified)
- **User**: name, student_id (unique, lowercased), phone, email, password hash, role.
- **Item**: type (lost/found), title, category, color, brand, location, date, description, images, approval_status, status.
- **Claim**: request_type (claim_request | found_match_request), status, admin_remarks.
- **Notification**: user_id, title, message, is_read.

---

## Folder map
- `main.go` – bootstrap, routing, template wiring, auto-migrations, data normalization.
- `internal/handler` – HTTP handlers (auth, item, admin, rendering helper).
- `internal/service` – business logic, validation, notifications, upload handling.
- `internal/repository` – DB operations (items, users, claims, notifications).
- `internal/middleware` – auth/session, admin guard, CSRF.
- `templates/` – HTML pages/partials.
- `static/` – CSS, JS, uploads saved to `static/uploads`.
- `pkg/database` – DB connection + masked logging.
- `pkg/utils` – bcrypt helpers.

---

## Deployment checklist
- Set `GO_ENV=production` and `COOKIE_SECURE=true` (HTTPS).
- Provide `SESSION_SECRET` ≥ 32 chars.
- Point to your PostgreSQL via `DATABASE_URL` or discrete DB_* vars.
- Rotate the seeded admin password immediately or create admins manually.
- Serve `static/uploads` with appropriate size/virus scanning policies if needed.

---

## Troubleshooting
- **No styles?** Ensure Tailwind CDN reachable; hard refresh. (Core layout still works with base CSS.)
- **DB connect errors?** Recheck env vars, network to Postgres, and that the DSN includes `sslmode`.
- **Upload failures?** Respect 5 MB/file and JPEG/PNG only; found posts require at least one photo.

---

## Contributing
1. Branch off `main`.
2. Keep server-rendered templates consistent; run `go test ./...`.
3. Avoid committing `.env` or real secrets; they’re intentionally gitignored.

