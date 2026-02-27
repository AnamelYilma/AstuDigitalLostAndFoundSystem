# ASTU Digital Lost & Found System Guide

## 1) What This System Does
This system is a web application for ASTU students and admins to manage lost and found items.

Core goals:
- Students can register, login, report lost/found items, search items, and claim found items.
- Admin can approve/reject/remove item posts and approve/reject claims.
- The system tracks approval and claim status.

## 2) Folder Structure (Current)

```text
AstuDigitalLostAndFoundSystem/
+-- main.go
+-- go.mod / go.sum
+-- pkg/
¦   +-- database/db.go
¦   +-- utils/hash.go
+-- internal/
¦   +-- model/
¦   ¦   +-- user.go
¦   ¦   +-- item.go
¦   +-- repository/
¦   ¦   +-- user_repository.go
¦   ¦   +-- item_repository.go
¦   +-- service/
¦   ¦   +-- auth_service.go
¦   ¦   +-- item_service.go
¦   ¦   +-- item_options.go
¦   +-- middleware/
¦   ¦   +-- auth_middleware.go
¦   +-- handler/
¦       +-- auth_handler.go
¦       +-- item_handler.go
¦       +-- admin_handler.go
+-- templates/
¦   +-- layout.html
¦   +-- index.html
¦   +-- login.html
¦   +-- register.html
¦   +-- dashboard.html
¦   +-- report.html
¦   +-- search.html
¦   +-- items.html
¦   +-- item.html
¦   +-- error.html
¦   +-- admin/
¦       +-- admin_dashboard.html
¦       +-- admin_claims.html
¦       +-- admin_items.html
+-- static/
    +-- css/style.css
    +-- uploads/
```

## 3) High-Level Architecture
The system uses layered architecture:

1. **Handler layer** (`internal/handler`)
- Reads HTTP request data.
- Calls services.
- Renders HTML templates or redirects.

2. **Service layer** (`internal/service`)
- Business rules and validation.
- Security checks (claim rules, upload validation, etc.).

3. **Repository layer** (`internal/repository`)
- All DB queries via GORM.
- Filtering, preload, update/delete logic.

4. **Model layer** (`internal/model`)
- DB table structures (`users`, `items`, `claims`).

5. **Middleware layer** (`internal/middleware`)
- Session auth checks.
- Admin authorization.
- Set current user in request context.

## 4) Request Flow (End-to-End)

```text
Browser -> Gin Router (main.go)
        -> Middleware (SetUser / AuthRequired / AdminRequired)
        -> Handler
        -> Service
        -> Repository (GORM)
        -> PostgreSQL
        -> Handler renders template (layout + content_template)
        -> Browser
```

## 5) Authentication and Session Flow
- Login uses `student_id + password`.
- Passwords are hashed with bcrypt (`pkg/utils/hash.go`).
- Session cookie stores `user_id`.
- On each request, `SetUser()` middleware loads user from DB and sets context.
- Protected pages use `AuthRequired()`.
- Admin pages use `AuthRequired()` + `AdminRequired()`.
- One login endpoint supports both roles:
  - `student` -> `/dashboard`
  - `admin` -> `/admin/dashboard`

## 6) Data Model and Status Meaning

### User
Fields include:
- `name`, `student_id`, `phone`, `email`, `password(hash)`, `role`

### Item
Fields include:
- `type`: `lost` or `found`
- `status`: item lifecycle (`open`, `claimed`)
- `approval_status`: moderation status (`pending`, `approved`, `rejected`)
- `admin_remarks`: admin feedback

### Claim
Fields include:
- `status`: `pending`, `approved`, `rejected`
- `admin_remarks`

## 7) Business Rules Implemented
- New reports are created with `approval_status = pending`.
- Claim can only be submitted when:
  - item type is `found`
  - item approval status is `approved`
  - item status is `open`
  - claimant is not the same user who posted item
- If claim approved, item status becomes `claimed`.
- Admin can approve/reject/remove item posts.

## 8) Search and Filter System
Search supports optional filters (single or combined):
- `type` (lost/found/all)
- `category`
- `location`
- `color`
- `approval status` (pending/approved/rejected/all)

Important behavior:
- If you leave filters empty and click Search, it returns all matching records by current query logic.
- Filters are independent and combinable.

## 9) Template Rendering System (HTML)
The project uses one base layout (`templates/layout.html`) and named content blocks.

Pattern:
- Layout checks `content_template` key.
- Handler sends `content_template` value.
- Matching template block is rendered.

Example:
- Handler sends `"content_template": "search_content"`
- Layout renders `{{ template "search_content" . }}`

This avoids duplicate layout code across pages.

## 10) Security Implementation
Already implemented:
- Password hashing (bcrypt).
- Role-based access control (admin routes protected).
- Session-based auth.
- Upload validation:
  - max 5MB
  - JPG/PNG only
- Input presence checks for core fields.

## 11) What You Need To Fix/Improve Next (Important)
To build a stronger production-ready system, improve these points:

1. Move DB credentials out of code
- `pkg/database/db.go` has hardcoded DSN.
- Read from `.env` instead.

2. Rotate session secret
- `auth_middleware.go` uses static secret string.
- Use environment variable and strong random secret.

3. Enforce DB constraints for `student_id`
- Ensure unique + non-null at DB migration level if business requires.

4. Add CSRF protection for POST forms
- Current forms are session-based but no CSRF token check.

5. Add server-side validation for category/type/other fields
- Validate against allowed constants consistently.

6. Add tests
- Unit tests for services.
- Integration tests for handlers and auth flow.

7. Clarify status naming
- You currently have both `status` and `approval_status`.
- Keep documented meanings consistent in UI labels.

8. Add audit logs
- Record admin moderation actions for accountability.

## 12) Build and Run Guide

### Prerequisites
- Go installed (see `go.mod` version target).
- PostgreSQL running.
- Database created (e.g., `lostfound`).

### Configure DB
Edit DSN in `pkg/database/db.go`:
- host
- user
- password
- dbname
- port

### Run
```bash
go mod tidy
go run main.go
```
Open:
- `http://localhost:8080`

### Default Admin
On first startup (if no admin exists), system creates:
- ID: `admin`
- Password: `admin123`

## 13) Typical User Flows

### Student flow
1. Register (`/register`).
2. Login (`/login`).
3. Report lost/found (`/report?type=lost` or `found`).
4. Search items (`/search`).
5. Open item detail (`/item/:id`).
6. Claim approved found item.
7. Track own post statuses in dashboard.

### Admin flow
1. Login from same login page using admin credentials.
2. Auto-redirect to `/admin/dashboard`.
3. Moderate item posts (`/admin/items`).
4. Approve/reject claims (`/admin/claims`).
5. View dashboard stats.

## 14) Notes About Current Files
- `templates/admin_login.html` currently exists but is not used by routes after one-login refactor.
- You can keep it for backup or remove it to reduce confusion.

---
If you want, I can create a second file `API_AND_ROUTE_MAP.md` with every route, method, middleware, and template mapping in a table.
