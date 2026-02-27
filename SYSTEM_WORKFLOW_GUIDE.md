# ASTU Digital Lost & Found System Guide

## 1) Purpose
This system helps ASTU students and admins manage lost and found items digitally.

Main goals:
- Students can register, login, report items, search items, and claim found items.
- Admin can approve/reject/remove item posts and approve/reject claims.
- Public explore/search shows only admin-approved posts.

## 2) Current Folder Structure
```text
AstuDigitalLostAndFoundSystem/
+-- main.go
+-- go.mod / go.sum
+-- pkg/
|   +-- database/db.go
|   +-- utils/hash.go
+-- internal/
|   +-- model/
|   |   +-- user.go
|   |   +-- item.go
|   +-- repository/
|   |   +-- user_repository.go
|   |   +-- item_repository.go
|   +-- service/
|   |   +-- auth_service.go
|   |   +-- item_service.go
|   |   +-- item_options.go
|   +-- middleware/
|   |   +-- auth_middleware.go
|   |   +-- csrf_middleware.go
|   +-- handler/
|       +-- auth_handler.go
|       +-- item_handler.go
|       +-- admin_handler.go
|       +-- render.go
+-- templates/
|   +-- layout.html
|   +-- index.html
|   +-- login.html
|   +-- register.html
|   +-- dashboard.html
|   +-- report.html
|   +-- search.html
|   +-- items.html
|   +-- item.html
|   +-- error.html
|   +-- admin/
|       +-- admin_dashboard.html
|       +-- admin_claims.html
|       +-- admin_items.html
+-- static/
    +-- css/style.css
    +-- uploads/
```

## 3) Architecture
The app uses layered architecture:
1. Handler layer: receives HTTP requests and renders templates.
2. Service layer: business rules and validations.
3. Repository layer: DB queries through GORM.
4. Model layer: table schemas (`users`, `items`, `claims`).
5. Middleware layer: auth, role, and CSRF checks.

## 4) Request Flow
```text
Browser -> Gin Router (main.go)
        -> Middleware (SetUser / CSRFMiddleware / AuthRequired / AdminRequired)
        -> Handler
        -> Service
        -> Repository (GORM)
        -> PostgreSQL
        -> Handler renders template (layout + content_template)
        -> Browser
```

## 5) Authentication and Roles
- Login uses `student_id + password`.
- Password is hashed with bcrypt.
- Session cookie stores `user_id`.
- One login endpoint is used by both student and admin.
- If role is `admin`, user is redirected to `/admin/dashboard`.

## 6) Data Model Summary
### User
- `name`, `student_id`, `phone`, `email`, `password`, `role`
- `student_id` is unique and not null.

### Item
- `type`: `lost` or `found`
- `status`: lifecycle (`open`, `claimed`)
- `approval_status`: moderation (`pending`, `approved`, `rejected`)
- `admin_remarks`

### Claim
- `status`: `pending`, `approved`, `rejected`
- `admin_remarks`

## 7) Business Rules
- New reported items are created with `approval_status = pending`.
- Non-admin search/explore only shows `approved` items.
- Non-admin users cannot open detail pages of unapproved items unless they are the owner.
- Claim submission is allowed only when:
  - item type is `found`
  - item approval is `approved`
  - item status is `open`
  - claimant is not the reporter
- If claim is approved, item status changes to `claimed`.

## 8) Search and Filter
Search supports single or combined optional filters:
- type (`lost`, `found`, or all)
- category
- location
- color
- approval status
- date range (`date_from`, `date_to`) using `created_at`

Filter behavior:
- Any filter can be used alone.
- Multiple filters can be combined.
- Empty filters show all records allowed by your role.

## 9) Template Rendering
The app uses one base layout (`templates/layout.html`) plus content blocks.

Pattern:
- Handler sets `content_template` value.
- Layout routes to matching template block.

Example:
- Handler sends `content_template = "search_content"`.
- Layout renders `{{ template "search_content" . }}`.

## 10) Security Controls Implemented
- Password hashing with bcrypt.
- Session-based authentication.
- Role-based authorization (admin-only routes protected).
- CSRF protection for `POST/PUT/PATCH/DELETE`.
- Input validation for type/category/location/status values.
- Upload validation:
  - max 5MB
  - JPG/PNG only
- Environment-based configuration for DB and session settings.

## 11) Environment Configuration
Use `.env` (or OS environment variables):
- `DB_HOST`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_PORT`
- `DB_SSLMODE`
- `SESSION_SECRET`
- `COOKIE_SECURE`

Start from `.env.example`.

## 12) Build and Run
1. Configure PostgreSQL and create DB (for example `lostfound`).
2. Configure `.env` values.
3. Run:
```bash
go mod tidy
go run main.go
```
4. Open `http://localhost:8080`.

## 13) Default Admin
If no admin exists, first startup creates:
- ID: `admin`
- Password: `admin123`

## 14) Optional Next Improvements
- Add unit and integration tests.
- Add audit log table for admin moderation actions.
- Add rate limiting for report/login endpoints.
- Improve status labels in UI for clarity (`status` vs `approval_status`).

## 15) Notes
- `templates/admin_login.html` exists but main routes currently use one combined login page (`/login`).
