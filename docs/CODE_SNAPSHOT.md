# Code Snapshot (All Code Files)

## Folder Map
`
.
├── 📁 docs
│   └── 📝 CODE_SNAPSHOT.md (this file)
├── 📁 internal
│   ├── 📁 handler
│   │   ├── 🐹 admin_handler.go
│   │   ├── 🐹 auth_handler.go
│   │   ├── 🐹 item_handler.go
│   │   └── 🐹 render.go
│   ├── 📁 middleware
│   │   ├── 🐹 auth_middleware.go
│   │   └── 🐹 csrf_middleware.go
│   ├── 📁 model
│   │   ├── 🐹 item.go
│   │   ├── 🐹 notification.go
│   │   └── 🐹 user.go
│   ├── 📁 repository
│   │   ├── 🐹 item_repository.go
│   │   └── 🐹 user_repository.go
│   └── 📁 service
│       ├── 🐹 auth_service.go
│       ├── 🐹 item_options.go
│       └── 🐹 item_service.go
├── 📁 pkg
│   ├── 📁 database
│   │   └── 🐹 db.go
│   └── 📁 utils
│       └── 🐹 hash.go
├── 📁 static
│   ├── 📁 css
│   │   └── 🎨 style.css
│   ├── 📁 img
│   │   └── 🖼️ logo.jpg.png
│   ├── 📁 js
│   │   └── 📄 image-upload.js
│   └── 📁 uploads (omitted)
├── 📁 templates
│   ├── 📁 admin
│   │   ├── 🌐 admin_claims.html
│   │   ├── 🌐 admin_dashboard.html
│   │   └── 🌐 admin_items.html
│   ├── 🌐 admin_login.html
│   ├── 🌐 dashboard.html
│   ├── 🌐 error.html
│   ├── 🌐 index.html
│   ├── 🌐 item.html
│   ├── 🌐 items.html
│   ├── 🌐 layout.html
│   ├── 🌐 login.html
│   ├── 🌐 notifications.html
│   ├── 🌐 register.html
│   └── 🌐 report.html
├── ⚙️ .air.toml
├── ⚙️ .gitattributes
├── ⚙️ .gitignore
├── 📝 ASTU STEAM COMMAND.md
├── 📝 README.md
├── 📄 go.mod
├── 📄 go.sum
└── 🐹 main.go
`

## File Contents

### main.go

`$lang
package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"lostfound/internal/handler"
	"lostfound/internal/middleware"
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"lostfound/pkg/utils"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
)

func main() {
	loadDotEnv(".env")
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	address := ":" + port

	if strings.EqualFold(os.Getenv("GO_ENV"), "production") {
		gin.SetMode(gin.ReleaseMode)
	}

	database.InitDB()
	database.DB.AutoMigrate(&model.User{}, &model.Item{}, &model.ItemImage{}, &model.Claim{}, &model.Notification{})
	createDefaultAdmin()
	normalizeLegacyData()
	enforceUserConstraints()

	r := gin.Default()
	funcMap := template.FuncMap{
		"now": func() string {
		return time.Now().Format("2006-01-02")
		},
	}

	tmpl := template.New("").Funcs(funcMap)

	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/admin/*.html"))

	r.SetHTMLTemplate(tmpl)



	r.StaticFS("/static", http.Dir("static"))
	r.Use(middleware.SetUser())
	r.Use(middleware.CSRFMiddleware())

	authHandler := handler.NewAuthHandler()
	itemHandler := handler.NewItemHandler()
	adminHandler := handler.NewAdminHandler()

	r.GET("/", func(c *gin.Context) {
		user, _ := c.Get("user")
		if u, ok := user.(model.User); ok {
			if u.Role == "admin" {
				c.Redirect(303, "/admin/dashboard")
				return
			}
			c.Redirect(303, "/dashboard")
			return
		}
		csrfToken, _ := c.Get("csrf_token")
		var unread int64
		c.HTML(200, "index.html", gin.H{
			"title":            "ASTU Lost & Found",
			"user":             user,
			"unread_count":     unread,
			"csrf_token":       csrfToken,
			"content_template": "index_content",
		})
	})

	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	r.GET("/register", authHandler.ShowRegister)
	r.POST("/register", authHandler.Register)
	r.GET("/logout", authHandler.Logout)

	r.GET("/report", itemHandler.Search)
	r.GET("/items", func(c *gin.Context) { c.Redirect(303, "/report") })
	r.GET("/search", func(c *gin.Context) { c.Redirect(303, "/report") })
	r.GET("/item/:id", itemHandler.ShowItem)

	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/dashboard", itemHandler.Dashboard)
		protected.GET("/report/new", itemHandler.ShowReportForm)
		protected.POST("/report/new", itemHandler.ReportItem)
		protected.POST("/claim", itemHandler.ClaimItem)
		protected.GET("/notifications", itemHandler.ShowNotifications)
		protected.POST("/notifications/read", itemHandler.MarkNotificationsRead)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/dashboard", adminHandler.Dashboard)
		admin.GET("/claims", adminHandler.ShowClaims)
		admin.POST("/claims/update", adminHandler.UpdateClaim)
		admin.GET("/items", adminHandler.ShowItems)
		admin.POST("/items/update", adminHandler.UpdateItem)
		admin.POST("/items/delete", adminHandler.DeleteItem)
	}


	log.Printf("Server starting on %s", address)
	if err := r.Run(address); err != nil { log.Fatal(err) }

}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func createDefaultAdmin() {
	var count int64
	database.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	if count == 0 {
		hashedPassword, _ := utils.HashPassword("admin123")
		admin := &model.User{
			Name:      "Admin",
			StudentID: "admin",
			Phone:     "0000000000",
			Email:     "admin@astu.edu",
			Password:  hashedPassword,
			Role:      "admin",
		}
		database.DB.Create(admin)
		log.Println("Default admin created: admin / admin123")
		return
	}

	var existingAdmin model.User
	if err := database.DB.Where("role = ?", "admin").First(&existingAdmin).Error; err == nil {
		needsUpdate := false
		if existingAdmin.StudentID == "" {
			existingAdmin.StudentID = "admin"
			needsUpdate = true
		}
		if existingAdmin.Phone == "" {
			existingAdmin.Phone = "0000000000"
			needsUpdate = true
		}
		if needsUpdate {
			database.DB.Save(&existingAdmin)
		}
	}
}

func normalizeLegacyData() {
	var users []model.User
	if err := database.DB.Order("id ASC").Find(&users).Error; err == nil {
		seen := map[string]bool{}
		for _, u := range users {
			needsUpdate := false

			sid := strings.ToLower(strings.TrimSpace(u.StudentID))
			if sid == "" {
				baseID := strings.Split(strings.TrimSpace(u.Email), "@")[0]
				if baseID == "" {
					baseID = fmt.Sprintf("user_%d", u.ID)
				}
				sid = strings.ToLower(baseID)
				needsUpdate = true
			}

			original := sid
			for seen[sid] {
				sid = fmt.Sprintf("%s_%d", original, u.ID)
				needsUpdate = true
			}
			seen[sid] = true

			if u.StudentID != sid {
				u.StudentID = sid
				needsUpdate = true
			}
			if strings.TrimSpace(u.Phone) == "" {
				u.Phone = "0000000000"
				needsUpdate = true
			}
			if needsUpdate {
				database.DB.Save(&u)
			}
		}
	}

	database.DB.Model(&model.Item{}).Where("approval_status IS NULL OR approval_status = ''").Update("approval_status", "approved")
	database.DB.Model(&model.Claim{}).Where("request_type IS NULL OR request_type = ''").Update("request_type", "claim_request")
}

func enforceUserConstraints() {
	// Keep old databases aligned with current model rules.
	database.DB.Exec("DROP INDEX IF EXISTS idx_users_student_id")
	database.DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_users_student_id_lower ON users (LOWER(student_id))")
	database.DB.Exec("ALTER TABLE users ALTER COLUMN student_id SET NOT NULL")
	database.DB.Exec("ALTER TABLE users ALTER COLUMN phone SET NOT NULL")
}


```

### .air.toml

`$lang
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main.exe ."
bin = "tmp/main.exe"
include_ext = ["go", "html", "css", "js", "env"]
include_dir = ["."]
exclude_dir = ["tmp", "vendor", ".git", "static/uploads"]
exclude_file = []
delay = 1000
stop_on_error = true
send_interrupt = true
kill_delay = 500

[log]
time = true
color = true

```

### .gitattributes

`$lang
# Auto detect text files and perform LF normalization
* text=auto

```

### .gitignore

`$lang
.env
tmp/
air.log

```

### ASTU STEAM COMMAND.md

`$lang
FINAL PROJECT - 1st & 2nd Year

Project Title: ASTU Digital Lost & Found System

Background & Problem Statement
Students at ASTU frequently lose personal belongings (ID cards, calculators, USB drives, lab coats, books, etc.).
This system provides a centralized digital platform to report, search, and manage lost/found items.

Project Objective
Build a secure platform to:
- Report lost items
- Report found items
- Search items
- Submit claims for found items
- Track approval workflow and status

System Roles
1. Student
- Register and login with Student ID + password
- Report lost/found item
- Search items (with or without login)
- Open item details and claim found items (login required)
- See own report status (pending/approved/rejected)

2. Admin
- Login from separate Admin Login page
- Approve/reject/remove item posts
- Approve/reject claim requests
- View user contact info for coordination
- View dashboard statistics

Campus Location List (ASTU)
- Library
- Cafe
- Class
- Lap
- Dorm
- On Road
- Tolest
- Shower
- Anphe
- Launch
- Park
- Hale.Birroe
- Other

Core Functionalities
- User registration (Name, Student ID, Phone, Password)
- Student login by Student ID + Password
- Separate admin login page
- Report lost/found with image upload
- Location selection from ASTU location list
- Search and filter by type, category, location, and color
- Public item search/list/detail available without login
- Claim submission requires login
- Admin approval workflow for item posts
- Claim approval/rejection workflow
- User dashboard showing approval + claim status

Workflow Rules
1. User submits item report -> status starts as pending
2. Admin approves/rejects/removes item report
3. Only approved items appear in public search
4. User can open large item detail without login
5. To apply/claim from detail page, user must login
6. User remains logged in until Logout is clicked

Security Requirements
- Password hashing
- Role-based access control
- Protected admin routes
- Input validation
- Secure upload handling
- Prevent unauthorized moderation/claim actions

Quick Run
1. Start PostgreSQL and ensure DB config in pkg/database/db.go is correct
2. Run: go run main.go
3. Open: http://localhost:8080

```

### README.md

`$lang
# ASTU Digital Lost & Found

Modern web app for reporting, browsing, and approving lost/found items at ASTU.

## Stack
- Go 1.21+, Gin web framework
- GORM + PostgreSQL
- HTML templates with Tailwind (CDN) + light custom CSS
- Gorilla sessions (cookie store), bcrypt password hashing
- CSRF middleware

## Folder Structure (key parts)
- `main.go` � bootstrap, routes, template loading, middleware
- `internal/handler` � HTTP handlers (auth, item, admin)
- `internal/service` � business rules (approvals, claims, notifications)
- `internal/repository` � DB access via GORM
- `internal/model` � User, Item, ItemImage, Claim, Notification structs
- `internal/middleware` � auth/session, admin guard, CSRF
- `templates` � HTML (layout, index, dashboard, report form, report view/cards, item detail, admin pages)
- `static/css/style.css` � minor overrides (Tailwind loaded via CDN)
- `static/uploads` � saved item images

## Run Locally
1) Set env (or `.env`):
```
DB_HOST=localhost
DB_PORT=000
DB_USER=postgres
DB_PASSWORD=yourpass
DB_NAME=NameDatabse
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

## System Flow (end-to-end)
1) **Report creation**
   - Student logs in ➜ opens `/report/new` (type=lost|found).
   - Submits details + optional photos (lost) / required photos (found).
   - Item stored as `pending` + `open`; invisible to public until approved.
2) **Admin review**
   - Admin dashboard lists pending posts.
   - Admin approves ➜ item becomes `approved` and shows on public Report View.
   - Admin rejects ➜ item stays hidden; poster sees remark.
3) **Public browsing**
   - Anyone can open `/report` and filter by type, category, location, color, date.
   - Only `approved` items are shown to guests; owners/admin can see their own pending items via dashboard.
4) **Requests / claims**
   - On item page:
     - If item type = **found** ➜ owner submits **Claim Request**.
     - If item type = **lost** ➜ finder submits **Found Match Request**.
   - Request goes to admin; item status stays `open`.
5) **Admin decision on requests**
   - Admin approves ➜ item status set to `claimed`; both users get notifications containing each other's contact (name, student ID, phone) to meet offline.
   - Admin rejects ➜ request closed; item remains `open`.
6) **Notifications**
   - Stored per user; badge count shown in navbar when logged in.
   - Users view/mark read at `/notifications`.

### Visibility & security rules
- Public list shows only `approved` items; contact info hidden until a request is approved.
- Only admins can approve/reject posts and requests.
- A user cannot request their own post.
- Image uploads validated (type/size); data validated against allowed enums (type/category/location/color).

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
- If Tailwind styles don�t show, hard-refresh (Ctrl+F5); templates load CDN script in `layout.html`.
- If DB fails: verify env vars and Postgres is running; check connection string.
- If build errors mention `search.html`: ensure `main.go` does not load removed templates (already adjusted).

## Development Tips
- Templates are server-rendered; update `templates/*.html` and restart `go run .`.
- Images are stored under `static/uploads`; ensure write perms.
- Use `go build ./...` to verify after edits.

```

### go.mod

`$lang
module lostfound

go 1.25.1

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/gorilla/sessions v1.4.0
	golang.org/x/crypto v0.48.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.54.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	go.uber.org/mock v0.5.0 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

```

### go.sum

`$lang
github.com/bytedance/sonic v1.14.0 h1:/OfKt8HFw0kh2rj8N0F6C/qPGRESq0BbaNZgcNXXzQQ=
github.com/bytedance/sonic v1.14.0/go.mod h1:WoEbx8WTcFJfzCe0hbmyTGrfjt8PzNEBdxlNUO24NhA=
github.com/bytedance/sonic/loader v0.3.0 h1:dskwH8edlzNMctoruo8FPTJDF3vLtDT0sXZwvZJyqeA=
github.com/bytedance/sonic/loader v0.3.0/go.mod h1:N8A3vUdtUebEY2/VQC0MyhYeKUFosQU6FxH2JmUe6VI=
github.com/cloudwego/base64x v0.1.6 h1:t11wG9AECkCDk5fMSoxmufanudBtJ+/HemLstXDLI2M=
github.com/cloudwego/base64x v0.1.6/go.mod h1:OFcloc187FXDaYHvrNIjxSe8ncn0OOM8gEHfghB2IPU=
github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/gabriel-vasile/mimetype v1.4.8 h1:FfZ3gj38NjllZIeJAmMhr+qKL8Wu+nOoI3GqacKw1NM=
github.com/gabriel-vasile/mimetype v1.4.8/go.mod h1:ByKUIKGjh1ODkGM1asKUbQZOLGrPjydw3hYPU2YU9t8=
github.com/gin-contrib/sse v1.1.0 h1:n0w2GMuUpWDVp7qSpvze6fAu9iRxJY4Hmj6AmBOU05w=
github.com/gin-contrib/sse v1.1.0/go.mod h1:hxRZ5gVpWMT7Z0B0gSNYqqsSCNIJMjzvm6fqCz9vjwM=
github.com/gin-gonic/gin v1.11.0 h1:OW/6PLjyusp2PPXtyxKHU0RbX6I/l28FTdDlae5ueWk=
github.com/gin-gonic/gin v1.11.0/go.mod h1:+iq/FyxlGzII0KHiBGjuNn4UNENUlKbGlNmc+W50Dls=
github.com/go-playground/assert/v2 v2.2.0 h1:JvknZsQTYeFEAhQwI4qEt9cyV5ONwRHC+lYKSsYSR8s=
github.com/go-playground/assert/v2 v2.2.0/go.mod h1:VDjEfimB/XKnb+ZQfWdccd7VUvScMdVu0Titje2rxJ4=
github.com/go-playground/locales v0.14.1 h1:EWaQ/wswjilfKLTECiXz7Rh+3BjFhfDFKv/oXslEjJA=
github.com/go-playground/locales v0.14.1/go.mod h1:hxrqLVvrK65+Rwrd5Fc6F2O76J/NuW9t0sjnWqG1slY=
github.com/go-playground/universal-translator v0.18.1 h1:Bcnm0ZwsGyWbCzImXv+pAJnYK9S473LQFuzCbDbfSFY=
github.com/go-playground/universal-translator v0.18.1/go.mod h1:xekY+UJKNuX9WP91TpwSH2VMlDf28Uj24BCp08ZFTUY=
github.com/go-playground/validator/v10 v10.27.0 h1:w8+XrWVMhGkxOaaowyKH35gFydVHOvC0/uWoy2Fzwn4=
github.com/go-playground/validator/v10 v10.27.0/go.mod h1:I5QpIEbmr8On7W0TktmJAumgzX4CA1XNl4ZmDuVHKKo=
github.com/goccy/go-json v0.10.2 h1:CrxCmQqYDkv1z7lO7Wbh2HN93uovUHgrECaO5ZrCXAU=
github.com/goccy/go-json v0.10.2/go.mod h1:6MelG93GURQebXPDq3khkgXZkazVtN9CRI+MGFi0w8I=
github.com/goccy/go-yaml v1.18.0 h1:8W7wMFS12Pcas7KU+VVkaiCng+kG8QiFeFwzFb+rwuw=
github.com/goccy/go-yaml v1.18.0/go.mod h1:XBurs7gK8ATbW4ZPGKgcbrY1Br56PdM69F7LkFRi1kA=
github.com/google/go-cmp v0.7.0 h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=
github.com/google/go-cmp v0.7.0/go.mod h1:pXiqmnSA92OHEEa9HXL2W4E7lf9JzCmGVUdgjX3N/iU=
github.com/google/gofuzz v1.0.0/go.mod h1:dBl0BpW6vV/+mYPU4Po3pmUjxk6FQPldtuIdl/M65Eg=
github.com/google/gofuzz v1.2.0 h1:xRy4A+RhZaiKjJ1bPfwQ8sedCA+YS2YcCHW6ec7JMi0=
github.com/google/gofuzz v1.2.0/go.mod h1:dBl0BpW6vV/+mYPU4Po3pmUjxk6FQPldtuIdl/M65Eg=
github.com/gorilla/securecookie v1.1.2 h1:YCIWL56dvtr73r6715mJs5ZvhtnY73hBvEF8kXD8ePA=
github.com/gorilla/securecookie v1.1.2/go.mod h1:NfCASbcHqRSY+3a8tlWJwsQap2VX5pwzwo4h3eOamfo=
github.com/gorilla/sessions v1.4.0 h1:kpIYOp/oi6MG/p5PgxApU8srsSw9tuFbt46Lt7auzqQ=
github.com/gorilla/sessions v1.4.0/go.mod h1:FLWm50oby91+hl7p/wRxDth9bWSuk0qVL2emc7lT5ik=
github.com/jackc/pgpassfile v1.0.0 h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=
github.com/jackc/pgpassfile v1.0.0/go.mod h1:CEx0iS5ambNFdcRtxPj5JhEz+xB6uRky5eyVu/W2HEg=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 h1:iCEnooe7UlwOQYpKFhBabPMi4aNAfoODPEFNiAnClxo=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761/go.mod h1:5TJZWKEWniPve33vlWYSoGYefn3gLQRzjfDlhSJ9ZKM=
github.com/jackc/pgx/v5 v5.6.0 h1:SWJzexBzPL5jb0GEsrPMLIsi/3jOo7RHlzTjcAeDrPY=
github.com/jackc/pgx/v5 v5.6.0/go.mod h1:DNZ/vlrUnhWCoFGxHAG8U2ljioxukquj7utPDgtQdTw=
github.com/jackc/puddle/v2 v2.2.2 h1:PR8nw+E/1w0GLuRFSmiioY6UooMp6KJv0/61nB7icHo=
github.com/jackc/puddle/v2 v2.2.2/go.mod h1:vriiEXHvEE654aYKXXjOvZM39qJ0q+azkZFrfEOc3H4=
github.com/jinzhu/inflection v1.0.0 h1:K317FqzuhWc8YvSVlFMCCUb36O/S9MCKRDI7QkRKD/E=
github.com/jinzhu/inflection v1.0.0/go.mod h1:h+uFLlag+Qp1Va5pdKtLDYj+kHp5pxUVkryuEj+Srlc=
github.com/jinzhu/now v1.1.5 h1:/o9tlHleP7gOFmsnYNz3RGnqzefHA47wQpKrrdTIwXQ=
github.com/jinzhu/now v1.1.5/go.mod h1:d3SSVoowX0Lcu0IBviAWJpolVfI5UJVZZ7cO71lE/z8=
github.com/json-iterator/go v1.1.12 h1:PV8peI4a0ysnczrg+LtxykD8LfKY9ML6u2jnxaEnrnM=
github.com/json-iterator/go v1.1.12/go.mod h1:e30LSqwooZae/UwlEbR2852Gd8hjQvJoHmT4TnhNGBo=
github.com/klauspost/cpuid/v2 v2.3.0 h1:S4CRMLnYUhGeDFDqkGriYKdfoFlDnMtqTiI/sFzhA9Y=
github.com/klauspost/cpuid/v2 v2.3.0/go.mod h1:hqwkgyIinND0mEev00jJYCxPNVRVXFQeu1XKlok6oO0=
github.com/leodido/go-urn v1.4.0 h1:WT9HwE9SGECu3lg4d/dIA+jxlljEa1/ffXKmRjqdmIQ=
github.com/leodido/go-urn v1.4.0/go.mod h1:bvxc+MVxLKB4z00jd1z+Dvzr47oO32F/QSNjSBOlFxI=
github.com/mattn/go-isatty v0.0.20 h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=
github.com/mattn/go-isatty v0.0.20/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 h1:ZqeYNhU3OHLH3mGKHDcjJRFFRrJa6eAM5H+CtDdOsPc=
github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421/go.mod h1:6dJC0mAP4ikYIbvyc7fijjWJddQyLn8Ig3JB5CqoB9Q=
github.com/modern-go/reflect2 v1.0.2 h1:xBagoLtFs94CBntxluKeaWgTMpvLxC4ur3nMaC9Gz0M=
github.com/modern-go/reflect2 v1.0.2/go.mod h1:yWuevngMOJpCy52FWWMvUC8ws7m/LJsjYzDa0/r8luk=
github.com/pelletier/go-toml/v2 v2.2.4 h1:mye9XuhQ6gvn5h28+VilKrrPoQVanw5PMw/TB0t5Ec4=
github.com/pelletier/go-toml/v2 v2.2.4/go.mod h1:2gIqNv+qfxSVS7cM2xJQKtLSTLUE9V8t9Stt+h56mCY=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/quic-go/qpack v0.5.1 h1:giqksBPnT/HDtZ6VhtFKgoLOWmlyo9Ei6u9PqzIMbhI=
github.com/quic-go/qpack v0.5.1/go.mod h1:+PC4XFrEskIVkcLzpEkbLqq1uCoxPhQuvK5rH1ZgaEg=
github.com/quic-go/quic-go v0.54.0 h1:6s1YB9QotYI6Ospeiguknbp2Znb/jZYjZLRXn9kMQBg=
github.com/quic-go/quic-go v0.54.0/go.mod h1:e68ZEaCdyviluZmy44P6Iey98v/Wfz6HCjQEm+l8zTY=
github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
github.com/stretchr/objx v0.4.0/go.mod h1:YvHI0jy2hoMjB+UWwv71VJQ9isScKT/TqJzVSSt89Yw=
github.com/stretchr/objx v0.5.0/go.mod h1:Yh+to48EsGEfYuaHDzXPcE3xhTkx73EhmCGUpEOglKo=
github.com/stretchr/testify v1.3.0/go.mod h1:M5WIy9Dh21IEIfnGCwXGc5bZfKNJtfHm1UVUgZn+9EI=
github.com/stretchr/testify v1.7.0/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
github.com/stretchr/testify v1.7.1/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
github.com/stretchr/testify v1.8.0/go.mod h1:yNjHg4UonilssWZ8iaSj1OCr/vHnekPRkoO+kdMU+MU=
github.com/stretchr/testify v1.8.1/go.mod h1:w2LPCIKwWwSfY2zedu0+kehJoqGctiVI29o6fzry7u4=
github.com/stretchr/testify v1.11.1 h1:7s2iGBzp5EwR7/aIZr8ao5+dra3wiQyKjjFuvgVKu7U=
github.com/stretchr/testify v1.11.1/go.mod h1:wZwfW3scLgRK+23gO65QZefKpKQRnfz6sD981Nm4B6U=
github.com/twitchyliquid64/golang-asm v0.15.1 h1:SU5vSMR7hnwNxj24w34ZyCi/FmDZTkS4MhqMhdFk5YI=
github.com/twitchyliquid64/golang-asm v0.15.1/go.mod h1:a1lVb/DtPvCB8fslRZhAngC2+aY1QWCk3Cedj/Gdt08=
github.com/ugorji/go/codec v1.3.0 h1:Qd2W2sQawAfG8XSvzwhBeoGq71zXOC/Q1E9y/wUcsUA=
github.com/ugorji/go/codec v1.3.0/go.mod h1:pRBVtBSKl77K30Bv8R2P+cLSGaTtex6fsA2Wjqmfxj4=
go.uber.org/mock v0.5.0 h1:KAMbZvZPyBPWgD14IrIQ38QCyjwpvVVV6K/bHl1IwQU=
go.uber.org/mock v0.5.0/go.mod h1:ge71pBPLYDk7QIi1LupWxdAykm7KIEFchiOqd6z7qMM=
golang.org/x/arch v0.20.0 h1:dx1zTU0MAE98U+TQ8BLl7XsJbgze2WnNKF/8tGp/Q6c=
golang.org/x/arch v0.20.0/go.mod h1:bdwinDaKcfZUGpH09BB7ZmOfhalA8lQdzl62l8gGWsk=
golang.org/x/crypto v0.48.0 h1:/VRzVqiRSggnhY7gNRxPauEQ5Drw9haKdM0jqfcCFts=
golang.org/x/crypto v0.48.0/go.mod h1:r0kV5h3qnFPlQnBSrULhlsRfryS2pmewsg+XfMgkVos=
golang.org/x/mod v0.32.0 h1:9F4d3PHLljb6x//jOyokMv3eX+YDeepZSEo3mFJy93c=
golang.org/x/mod v0.32.0/go.mod h1:SgipZ/3h2Ci89DlEtEXWUk/HteuRin+HHhN+WbNhguU=
golang.org/x/net v0.49.0 h1:eeHFmOGUTtaaPSGNmjBKpbng9MulQsJURQUAfUwY++o=
golang.org/x/net v0.49.0/go.mod h1:/ysNB2EvaqvesRkuLAyjI1ycPZlQHM3q01F02UY/MV8=
golang.org/x/sync v0.19.0 h1:vV+1eWNmZ5geRlYjzm2adRgW2/mcpevXNg50YZtPCE4=
golang.org/x/sync v0.19.0/go.mod h1:9KTHXmSnoGruLpwFjVSX0lNNA75CykiMECbovNTZqGI=
golang.org/x/sys v0.6.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.41.0 h1:Ivj+2Cp/ylzLiEU89QhWblYnOE9zerudt9Ftecq2C6k=
golang.org/x/sys v0.41.0/go.mod h1:OgkHotnGiDImocRcuBABYBEXf8A9a87e/uXjp9XT3ks=
golang.org/x/text v0.34.0 h1:oL/Qq0Kdaqxa1KbNeMKwQq0reLCCaFtqu2eNuSeNHbk=
golang.org/x/text v0.34.0/go.mod h1:homfLqTYRFyVYemLBFl5GgL/DWEiH5wcsQ5gSh1yziA=
golang.org/x/tools v0.41.0 h1:a9b8iMweWG+S0OBnlU36rzLp20z1Rp10w+IY2czHTQc=
golang.org/x/tools v0.41.0/go.mod h1:XSY6eDqxVNiYgezAVqqCeihT4j1U2CCsqvH3WhQpnlg=
google.golang.org/protobuf v1.36.9 h1:w2gp2mA27hUeUzj9Ex9FBjsBm40zfaDtEWow293U7Iw=
google.golang.org/protobuf v1.36.9/go.mod h1:fuxRtAxBytpl4zzqUh6/eyUujkJdNiuEkXntxiD/uRU=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
gorm.io/driver/postgres v1.6.0 h1:2dxzU8xJ+ivvqTRph34QX+WrRaJlmfyPqXmoGVjMBa4=
gorm.io/driver/postgres v1.6.0/go.mod h1:vUw0mrGgrTK+uPHEhAdV4sfFELrByKVGnaVRkXDhtWo=
gorm.io/gorm v1.31.1 h1:7CA8FTFz/gRfgqgpeKIBcervUn3xSyPUmr6B2WXJ7kg=
gorm.io/gorm v1.31.1/go.mod h1:XyQVbO2k6YkOis7C2437jSit3SsDK72s7n7rsSHd+Gs=

```

### internal/handler/admin_handler.go

`$lang
package handler

import (
	"lostfound/internal/model"
	"lostfound/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	itemService *service.ItemService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		itemService: service.NewItemService(),
	}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	user, _ := c.Get("user")
	stats, _ := h.itemService.GetStats()
	claims, _ := h.itemService.GetAllClaims()

	renderHTML(c, http.StatusOK, "admin_dashboard.html", gin.H{
		"title":            "Admin Dashboard",
		"user":             user,
		"stats":            stats,
		"claims":           claims,
		"content_template": "admin_dashboard_content",
	})
}

func (h *AdminHandler) ShowClaims(c *gin.Context) {
	user, _ := c.Get("user")
	claims, _ := h.itemService.GetAllClaims()

	renderHTML(c, http.StatusOK, "admin_claims.html", gin.H{
		"title":            "Manage Claims",
		"user":             user,
		"claims":           claims,
		"content_template": "admin_claims_content",
	})
}

func (h *AdminHandler) ShowItems(c *gin.Context) {
	user, _ := c.Get("user")
	filters := map[string]interface{}{}
	selectedQ := strings.TrimSpace(c.Query("q"))
	selectedType := ""
	selectedStatus := ""
	selectedCategory := ""
	selectedLocation := ""
	selectedDateFrom := ""
	selectedDateTo := ""

	if selectedQ != "" {
		filters["q"] = selectedQ
	}
	if t := c.Query("type"); t != "" && service.IsValidItemType(t) {
		filters["type"] = t
		selectedType = t
	}
	if st := c.Query("status"); st != "" && service.IsValidApprovalStatus(st) {
		filters["approval_status"] = st
		selectedStatus = st
	}
	if cgy := c.Query("category"); cgy != "" && service.IsValidCategory(cgy) {
		filters["category"] = cgy
		selectedCategory = cgy
	}
	if loc := c.Query("location"); loc != "" && service.IsValidASTULocation(loc) {
		filters["location"] = loc
		selectedLocation = loc
	}
	if from := c.Query("date_from"); from != "" {
		if _, err := time.Parse("2006-01-02", from); err == nil {
			filters["date_from"] = from
			selectedDateFrom = from
		}
	}
	if to := c.Query("date_to"); to != "" {
		if _, err := time.Parse("2006-01-02", to); err == nil {
			filters["date_to"] = to
			selectedDateTo = to
		}
	}

	items, _ := h.itemService.SearchItems(filters)
	typedUser, _ := user.(model.User)
	unreadCount := h.itemService.CountUnreadNotifications(typedUser.ID)

	renderHTML(c, http.StatusOK, "admin_items.html", gin.H{
		"title":              "Manage Item Posts",
		"user":               user,
		"items":              items,
		"filters":            filters,
		"locations":          service.ASTULocations(),
		"categories":         service.ItemCategories(),
		"selected_q":         selectedQ,
		"selected_type":      selectedType,
		"selected_status":    selectedStatus,
		"selected_category":  selectedCategory,
		"selected_location":  selectedLocation,
		"selected_date_from": selectedDateFrom,
		"selected_date_to":   selectedDateTo,
		"unread_count":       unreadCount,
		"content_template":   "admin_items_content",
	})
}

func (h *AdminHandler) UpdateClaim(c *gin.Context) {
	claimID, _ := strconv.ParseUint(c.PostForm("claim_id"), 10, 32)
	status := c.PostForm("status")
	remarks := c.PostForm("remarks")

	err := h.itemService.UpdateClaimStatus(uint(claimID), status, remarks)
	if err != nil {
		renderHTML(c, http.StatusOK, "error.html", gin.H{
			"error":            "Failed to update claim: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/claims")
}

func (h *AdminHandler) UpdateItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.PostForm("item_id"), 10, 32)
	status := c.PostForm("approval_status")
	remarks := c.PostForm("remarks")

	err := h.itemService.UpdateItemApproval(uint(itemID), status, remarks)
	if err != nil {
		renderHTML(c, http.StatusOK, "error.html", gin.H{
			"error":            "Failed to update item status: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/items")
}

func (h *AdminHandler) DeleteItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.PostForm("item_id"), 10, 32)
	err := h.itemService.DeleteItem(uint(itemID))
	if err != nil {
		renderHTML(c, http.StatusOK, "error.html", gin.H{
			"error":            "Failed to remove item: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/items")
}

```

### internal/handler/auth_handler.go

`$lang
package handler

import (
	"lostfound/internal/middleware"
	"lostfound/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	renderHTML(c, http.StatusOK, "login.html", gin.H{
		"title":            "Login",
		"content_template": "login_content",
	})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	renderHTML(c, http.StatusOK, "register.html", gin.H{
		"title":            "Register",
		"content_template": "register_content",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	studentID := c.PostForm("student_id")
	password := c.PostForm("password")

	user, err := h.authService.Login(studentID, password)
	if err != nil {
		renderHTML(c, http.StatusOK, "login.html", gin.H{
			"title":            "Login",
			"error":            err.Error(),
			"content_template": "login_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	if user.Role == "admin" {
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	} else {
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	name := c.PostForm("name")
	studentID := c.PostForm("student_id")
	phone := c.PostForm("phone")
	password := c.PostForm("password")

	user, err := h.authService.Register(name, studentID, phone, password)
	if err != nil {
		renderHTML(c, http.StatusOK, "register.html", gin.H{
			"title":            "Register",
			"error":            err.Error(),
			"content_template": "register_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := middleware.GetSession(c)
	session.Values = make(map[interface{}]interface{})
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusSeeOther, "/")
}

```

### internal/handler/item_handler.go

`$lang
package handler

import (
	"lostfound/internal/model"
	"net/http"
	"strconv"
	"strings"
	"time"
	// "lostfound/internal/middleware"
	"lostfound/internal/service"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	itemService *service.ItemService
}

func NewItemHandler() *ItemHandler {
	return &ItemHandler{
		itemService: service.NewItemService(),
	}
}

func (h *ItemHandler) Dashboard(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	stats, _ := h.itemService.GetStats()
	myItems, _ := h.itemService.GetItemsByUserID(user.ID)
	unreadCount := h.itemService.CountUnreadNotifications(user.ID)

	renderHTML(c, http.StatusOK, "dashboard.html", gin.H{
		"title":            "Dashboard",
		"user":             user,
		"stats":            stats,
		"my_items":         myItems,
		"unread_count":     unreadCount,
		"content_template": "dashboard_content",
	})
}

func (h *ItemHandler) ShowReportForm(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	itemType := c.Query("type")
	if itemType != "lost" && itemType != "found" {
		itemType = "lost"
	}
	categories := service.ItemCategories()
	locations := service.ASTULocations()
	colors := service.ColorOptions()

	renderHTML(c, http.StatusOK, "report.html", gin.H{
		"title":            "Report Item",
		"user":             user,
		"type":             itemType,
		"locations":        locations,
		"colors":           colors,
		"categories":       categories,
		"content_template": "report_content",
	})
}

func (h *ItemHandler) ReportItem(c *gin.Context) {
	u := c.MustGet("user").(model.User)
	locations := service.ASTULocations()
	colors := service.ColorOptions()
	categories := service.ItemCategories()

	itemType := c.PostForm("type")
	title := c.PostForm("title")
	category := c.PostForm("category")
	color := c.PostForm("color")
	if strings.EqualFold(color, "other") {
		color = strings.TrimSpace(c.PostForm("color_other"))
	}
	color = strings.ToLower(strings.TrimSpace(color))
	brand := c.PostForm("brand")
	location := c.PostForm("location")
	date := c.PostForm("date")
	description := c.PostForm("description")
	if !service.IsValidItemType(itemType) {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             "lost",
			"error":            "Invalid report type",
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}
	if !service.IsValidCategory(category) {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Invalid category value",
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}
	if !service.IsValidASTULocation(location) {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Please select a valid ASTU location",
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}
	if strings.TrimSpace(color) == "" {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Please provide item color",
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}

	var imagePaths []string
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["images"]
		for _, file := range files {
			path, saveErr := h.itemService.SaveImage(file)
			if saveErr != nil {
				renderHTML(c, http.StatusOK, "report.html", gin.H{
					"title":            "Report Item",
					"user":             u,
					"type":             itemType,
					"error":            "Image upload failed: " + saveErr.Error(),
					"locations":        locations,
					"colors":           colors,
					"categories":       categories,
					"content_template": "report_content",
				})
				return
			}
			imagePaths = append(imagePaths, path)
		}
	}
	if len(imagePaths) == 0 {
		// Backward compatibility for single-image field name
		if file, err := c.FormFile("image"); err == nil {
			path, saveErr := h.itemService.SaveImage(file)
			if saveErr != nil {
				renderHTML(c, http.StatusOK, "report.html", gin.H{
					"title":            "Report Item",
					"user":             u,
					"type":             itemType,
					"error":            "Image upload failed: " + saveErr.Error(),
					"locations":        locations,
					"colors":           colors,
					"categories":       categories,
					"content_template": "report_content",
				})
				return
			}
			imagePaths = append(imagePaths, path)
		}
	}
	if len(imagePaths) == 0 && itemType == "found" {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Photo is required for found item reports",
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}

	_, err := h.itemService.CreateItem(
		u.ID, itemType, title, category, color, brand,
		location, date, description, imagePaths,
	)

	if err != nil {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Failed to save item: " + err.Error(),
			"locations":        locations,
			"colors":           colors,
			"categories":       categories,
			"content_template": "report_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ItemHandler) ShowSearch(c *gin.Context) {
	user, _ := c.Get("user")
	isAdmin := false
	if user != nil {
		u := user.(model.User)
		isAdmin = u.Role == "admin"
	}
	renderHTML(c, http.StatusOK, "search.html", gin.H{
		"title":            "Search Items",
		"user":             user,
		"is_admin":         isAdmin,
		"locations":        service.ASTULocations(),
		"colors":           service.ColorOptions(),
		"content_template": "search_content",
	})
}

func (h *ItemHandler) Search(c *gin.Context) {
	filters := make(map[string]interface{})
	user, _ := c.Get("user")
	isAdmin := false
	if user != nil {
		u := user.(model.User)
		isAdmin = u.Role == "admin"
	}

	selectedCategory := ""
	selectedLocation := ""
	selectedType := ""
	selectedColor := ""
	selectedDateFrom := ""
	selectedDateTo := ""

	if category := c.Query("category"); category != "" {
		if service.IsValidCategory(category) {
			filters["category"] = category
			selectedCategory = category
		}
	}
	if location := c.Query("location"); location != "" {
		if service.IsValidASTULocation(location) {
			filters["location"] = location
			selectedLocation = location
		}
	}
	var selectedColors []string
	for _, color := range c.QueryArray("color") {
		clean := strings.ToLower(strings.TrimSpace(color))
		if clean == "" {
			continue
		}
		if service.IsStandardColor(clean) {
			selectedColors = append(selectedColors, clean)
		}
	}
	if len(selectedColors) > 0 {
		filters["colors"] = selectedColors
		if len(selectedColors) == 1 {
			selectedColor = selectedColors[0]
		}
	}
	if itemType := c.Query("type"); itemType != "" {
		if service.IsValidItemType(itemType) {
			filters["type"] = itemType
			selectedType = itemType
		}
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = dateFrom
			selectedDateFrom = dateFrom
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = dateTo
			selectedDateTo = dateTo
		}
	}
	// Default: only approved posts for everyone.
	filters["approval_status"] = "approved"

	// Admin can override via query param.
	if status := c.Query("status"); status != "" && isAdmin {
		if service.IsValidApprovalStatus(status) {
			filters["approval_status"] = status
		}
	}

	items, err := h.itemService.SearchItems(filters)
	if err != nil {
		items = []model.Item{}
	}

	renderHTML(c, http.StatusOK, "items.html", gin.H{
		"title":              "Report View",
		"items":              items,
		"filters":            filters,
		"selected_colors":    selectedColors,
		"user":               user,
		"is_admin":           isAdmin,
		"locations":          service.ASTULocations(),
		"colors":             service.ColorOptions(),
		"categories":         service.ItemCategories(),
		"selected_category":  selectedCategory,
		"selected_location":  selectedLocation,
		"selected_type":      selectedType,
		"selected_color":     selectedColor,
		"selected_date_from": selectedDateFrom,
		"selected_date_to":   selectedDateTo,
		"content_template":   "items_content",
	})
}

func (h *ItemHandler) ShowItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	item, err := h.itemService.GetItemByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/report")
		return
	}

	user, _ := c.Get("user")
	showPrivateInfo := false
	canRequest := false
	requestTypeLabel := "Claim Request"
	requestDescriptionHint := "Why do you think this item belongs to you?"
	if item.Type == "lost" {
		requestTypeLabel = "Found Match Request"
		requestDescriptionHint = "Describe where/when you found this item and proof details."
	}

	if item.ApprovalStatus != "approved" {
		if user == nil {
			c.Redirect(http.StatusSeeOther, "/report")
			return
		}
		u := user.(model.User)
		if u.Role != "admin" && u.ID != item.UserID {
			c.Redirect(http.StatusSeeOther, "/report")
			return
		}
	}
	if user != nil {
		u := user.(model.User)
		canRequest = (u.ID != item.UserID) && item.ApprovalStatus == "approved" && item.Status == "open"
		if u.Role == "admin" || u.ID == item.UserID || h.itemService.HasApprovedRequestForUser(item.ID, u.ID) {
			showPrivateInfo = true
		}
	}

	renderHTML(c, http.StatusOK, "item.html", gin.H{
		"title":            item.Title,
		"item":             item,
		"user":             user,
		"can_request":      canRequest,
		"show_private":     showPrivateInfo,
		"request_type":     requestTypeLabel,
		"request_hint":     requestDescriptionHint,
		"locations":        service.ASTULocations(),
		"categories":       service.ItemCategories(),
		"colors":           service.ColorOptions(),
		"content_template": "item_content",
	})
}

func (h *ItemHandler) ClaimItem(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(model.User)

	itemID, _ := strconv.ParseUint(c.PostForm("item_id"), 10, 32)
	description := strings.TrimSpace(c.PostForm("description"))

	// Optional structured details to help admin match without exposing on cards.
	claimLocation := strings.TrimSpace(c.PostForm("claim_location"))
	claimCategory := strings.TrimSpace(c.PostForm("claim_category"))
	claimColor := strings.TrimSpace(c.PostForm("claim_color"))
	claimDate := strings.TrimSpace(c.PostForm("claim_date"))

	parts := []string{}
	if claimLocation != "" {
		parts = append(parts, "Location: "+claimLocation)
	}
	if claimCategory != "" {
		parts = append(parts, "Category: "+claimCategory)
	}
	if claimColor != "" {
		parts = append(parts, "Color: "+claimColor)
	}
	if claimDate != "" {
		parts = append(parts, "Date: "+claimDate)
	}
	if description != "" {
		parts = append(parts, "Notes: "+description)
	}
	if len(parts) > 0 {
		description = strings.Join(parts, " | ")
	}

	err := h.itemService.CreateClaim(uint(itemID), u.ID, description)
	if err != nil {
		renderHTML(c, http.StatusOK, "error.html", gin.H{
			"error":            "Failed to submit request: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ItemHandler) ShowNotifications(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	notifications, _ := h.itemService.GetNotificationsByUserID(user.ID)

	renderHTML(c, http.StatusOK, "notifications.html", gin.H{
		"title":            "Notifications",
		"user":             user,
		"notifications":    notifications,
		"content_template": "notifications_content",
	})
}

func (h *ItemHandler) MarkNotificationsRead(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	_ = h.itemService.MarkNotificationsRead(user.ID)
	c.Redirect(http.StatusSeeOther, "/notifications")
}

```

### internal/handler/render.go

`$lang
package handler

import (
	"lostfound/internal/model"
	"lostfound/internal/service"
	"lostfound/pkg/database"

	"github.com/gin-gonic/gin"
)

func renderHTML(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	if token, ok := c.Get("csrf_token"); ok {
		if _, exists := data["csrf_token"]; !exists {
			data["csrf_token"] = token
		}
	}
	if userVal, ok := c.Get("user"); ok {
		if _, exists := data["user"]; !exists {
			data["user"] = userVal
		}
		if u, ok := userVal.(model.User); ok {
			if _, exists := data["unread_count"]; !exists {
				var unread int64
				database.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", u.ID, false).Count(&unread)
				data["unread_count"] = unread
			}
		}
	}
	// Provide common option lists if a template forgets to set them.
	if _, exists := data["locations"]; !exists {
		data["locations"] = service.ASTULocations()
	}
	if _, exists := data["categories"]; !exists {
		data["categories"] = service.ItemCategories()
	}
	if _, exists := data["colors"]; !exists {
		data["colors"] = service.ColorOptions()
	}
	c.HTML(status, name, data)
}

```

### internal/middleware/auth_middleware.go

`$lang
package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var (
	store     *sessions.CookieStore
	storeOnce sync.Once
)

func initStore() {
	store = sessions.NewCookieStore([]byte(getSessionSecret()))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}
}

func getSessionSecret() string {
	if v := os.Getenv("SESSION_SECRET"); v != "" {
		if len(v) >= 32 {
			return v
		}
		sum := sha256.Sum256([]byte(v))
		log.Println("SESSION_SECRET is shorter than 32 chars; using SHA-256 derived key")
		return base64.StdEncoding.EncodeToString(sum[:])
	}
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err == nil {
		log.Println("SESSION_SECRET not set; generated an ephemeral secret for this run")
		return base64.StdEncoding.EncodeToString(buf)
	}
	log.Println("SESSION_SECRET not set and secure random failed; using fallback development secret")
	return "dev-only-session-secret-change-this-32chars"
}

func GetSession(c *gin.Context) *sessions.Session {
	storeOnce.Do(initStore)
	session, _ := store.Get(c.Request, "auth-session")
	return session
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		userID, ok := session.Values["user_id"]

		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		var user model.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			session.Values = make(map[interface{}]interface{})
			session.Save(c.Request, c.Writer)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		u := user.(model.User)
		if u.Role != "admin" {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error":            "Access denied. Admin only.",
				"title":            "Forbidden",
				"content_template": "error_content",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		userID, ok := session.Values["user_id"]

		if ok {
			var user model.User
			if err := database.DB.First(&user, userID).Error; err == nil {
				c.Set("user", user)
			}
		}

		c.Next()
	}
}

```

### internal/middleware/csrf_middleware.go

`$lang
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfSessionKey = "csrf_token"

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		token, _ := session.Values[csrfSessionKey].(string)
		if token == "" {
			token = randomToken(32)
			session.Values[csrfSessionKey] = token
			_ = session.Save(c.Request, c.Writer)
		}

		c.Set("csrf_token", token)

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			sent := c.PostForm("csrf_token")
			if sent == "" {
				sent = c.GetHeader("X-CSRF-Token")
			}
			if sent == "" || subtle.ConstantTimeCompare([]byte(sent), []byte(token)) != 1 {
				c.HTML(http.StatusForbidden, "error.html", gin.H{
					"title":            "Security Error",
					"error":            "Invalid CSRF token. Refresh and try again.",
					"content_template": "error_content",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

```

### internal/model/item.go

`$lang
package model

import (
	"gorm.io/gorm"
	"time"
)

type Item struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	Type           string         `gorm:"size:10;not null" json:"type"`
	Title          string         `gorm:"size:200;not null" json:"title"`
	Category       string         `gorm:"size:50;not null" json:"category"`
	Color          string         `gorm:"size:50" json:"color"`
	Brand          string         `gorm:"size:100" json:"brand"`
	Location       string         `gorm:"size:200;not null" json:"location"`
	Date           string         `gorm:"size:20" json:"date"`
	Description    string         `gorm:"type:text" json:"description"`
	Image          string         `gorm:"size:500" json:"image"`
	Status         string         `gorm:"size:20;default:'open'" json:"status"`
	ApprovalStatus string         `gorm:"size:20;default:'pending'" json:"approval_status"`
	AdminRemarks   string         `gorm:"type:text" json:"admin_remarks"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	User   User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Images []ItemImage `gorm:"foreignKey:ItemID" json:"images,omitempty"`
}

type ItemImage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ItemID    uint           `gorm:"index;not null" json:"item_id"`
	Path      string         `gorm:"size:500;not null" json:"path"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Claim struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ItemID       uint           `gorm:"not null" json:"item_id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	RequestType  string         `gorm:"size:30;default:'claim_request'" json:"request_type"`
	Description  string         `gorm:"type:text" json:"description"`
	Status       string         `gorm:"size:20;default:'pending'" json:"status"`
	AdminRemarks string         `gorm:"type:text" json:"admin_remarks"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Item Item `gorm:"foreignKey:ItemID" json:"item"`
	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (Item) TableName() string {
	return "items"
}

func (Claim) TableName() string {
	return "claims"
}

func (ItemImage) TableName() string {
	return "item_images"
}

```

### internal/model/notification.go

`$lang
package model

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Title     string         `gorm:"size:200;not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}

```

### internal/model/user.go

`$lang
package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	StudentID string         `gorm:"size:30;uniqueIndex;not null" json:"student_id"`
	Phone     string         `gorm:"size:20;not null" json:"phone"`
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:20;default:'student'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

```

### internal/repository/item_repository.go

`$lang
package repository

import (
	"lostfound/internal/model"
	"lostfound/pkg/database"
)

type ItemRepository struct{}

func NewItemRepository() *ItemRepository {
	return &ItemRepository{}
}

func (r *ItemRepository) Create(item *model.Item) error {
	return database.DB.Create(item).Error
}

func (r *ItemRepository) FindAll(filters map[string]interface{}) ([]model.Item, error) {
	var items []model.Item
	query := database.DB.Preload("User").Preload("Images")

	if approvalStatus, ok := filters["approval_status"]; ok && approvalStatus != "" {
		query = query.Where("approval_status = ?", approvalStatus)
	}

	if category, ok := filters["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if location, ok := filters["location"]; ok && location != "" {
		query = query.Where("location = ?", location)
	}
	if colors, ok := filters["colors"]; ok {
		if colorList, ok := colors.([]string); ok && len(colorList) > 0 {
			knownColors := []string{"red", "green", "blue", "yellow", "black", "white", "gray", "brown", "orange", "purple", "pink", "gold", "silver"}
			hasOther := false
			standardSelected := make([]string, 0, len(colorList))
			for _, c := range colorList {
				if c == "other" {
					hasOther = true
					continue
				}
				standardSelected = append(standardSelected, c)
			}

			switch {
			case hasOther && len(standardSelected) > 0:
				query = query.Where("(LOWER(color) IN ? OR (LOWER(color) NOT IN ? AND TRIM(color) <> ''))", standardSelected, knownColors)
			case hasOther:
				query = query.Where("LOWER(color) NOT IN ? AND TRIM(color) <> ''", knownColors)
			default:
				query = query.Where("LOWER(color) IN ?", standardSelected)
			}
		}
	}
	
	if itemType, ok := filters["type"]; ok && itemType != "" {
		query = query.Where("type = ?", itemType)
	}
	if q, ok := filters["q"]; ok {
		if keyword, ok := q.(string); ok && keyword != "" {
			like := "%" + keyword + "%"
			query = query.Joins("LEFT JOIN users ON users.id = items.user_id").Where(
				"LOWER(items.title) LIKE LOWER(?) OR LOWER(items.description) LIKE LOWER(?) OR LOWER(items.brand) LIKE LOWER(?) OR LOWER(users.name) LIKE LOWER(?) OR LOWER(users.student_id) LIKE LOWER(?)",
				like, like, like, like, like,
			)
		}
	}
	if dateFrom, ok := filters["date_from"]; ok && dateFrom != "" {
		query = query.Where("\"date\" >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"]; ok && dateTo != "" {
		query = query.Where("\"date\" <= ?", dateTo)
	}

	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindByID(id uint) (*model.Item, error) {
	var item model.Item
	err := database.DB.Preload("User").Preload("Images").First(&item, id).Error
	return &item, err
}

func (r *ItemRepository) Update(item *model.Item) error {
	return database.DB.Save(item).Error
}

func (r *ItemRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	var totalLost int64
	var totalFound int64
	var totalClaims int64
	var pendingClaims int64
	var pendingItems int64

	database.DB.Model(&model.Item{}).Where("type = ? AND approval_status = ?", "lost", "approved").Count(&totalLost)
	database.DB.Model(&model.Item{}).Where("type = ? AND approval_status = ?", "found", "approved").Count(&totalFound)
	database.DB.Model(&model.Claim{}).Count(&totalClaims)
	database.DB.Model(&model.Claim{}).Where("status = ?", "pending").Count(&pendingClaims)
	database.DB.Model(&model.Item{}).Where("approval_status = ?", "pending").Count(&pendingItems)

	stats["total_lost"] = totalLost
	stats["total_found"] = totalFound
	stats["total_claims"] = totalClaims
	stats["pending_claims"] = pendingClaims
	stats["pending_items"] = pendingItems

	return stats, nil
}

func (r *ItemRepository) CreateClaim(claim *model.Claim) error {
	return database.DB.Create(claim).Error
}

func (r *ItemRepository) FindAllClaims() ([]model.Claim, error) {
	var claims []model.Claim
	err := database.DB.Preload("Item").Preload("Item.User").Preload("User").Order("created_at DESC").Find(&claims).Error
	return claims, err
}

func (r *ItemRepository) UpdateClaim(claim *model.Claim) error {
	return database.DB.Save(claim).Error
}

func (r *ItemRepository) FindClaimByID(id uint) (*model.Claim, error) {
	var claim model.Claim
	err := database.DB.Preload("Item").Preload("Item.User").Preload("User").First(&claim, id).Error
	return &claim, err
}

func (r *ItemRepository) HasActiveClaimByUser(itemID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Claim{}).
		Where("item_id = ? AND user_id = ? AND status IN ?", itemID, userID, []string{"pending", "approved"}).
		Count(&count).Error
	return count > 0, err
}

func (r *ItemRepository) HasApprovedClaimForUser(itemID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Claim{}).
		Where("item_id = ? AND user_id = ? AND status = ?", itemID, userID, "approved").
		Count(&count).Error
	return count > 0, err
}

func (r *ItemRepository) FindByUserID(userID uint) ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Preload("Images").Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindAllItemsForAdmin() ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Preload("User").Preload("Images").Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) DeleteItem(itemID uint) error {
	return database.DB.Delete(&model.Item{}, itemID).Error
}

func (r *ItemRepository) CreateItemImages(itemID uint, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	images := make([]model.ItemImage, 0, len(paths))
	for _, path := range paths {
		images = append(images, model.ItemImage{
			ItemID: itemID,
			Path:   path,
		})
	}
	return database.DB.Create(&images).Error
}

func (r *ItemRepository) CreateNotification(notification *model.Notification) error {
	return database.DB.Create(notification).Error
}

func (r *ItemRepository) FindNotificationsByUserID(userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *ItemRepository) MarkNotificationsRead(userID uint) error {
	return database.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

func (r *ItemRepository) CountUnreadNotifications(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *ItemRepository) FindAdmins() ([]model.User, error) {
	var admins []model.User
	err := database.DB.Where("role = ?", "admin").Find(&admins).Error
	return admins, err
}

```

### internal/repository/user_repository.go

`$lang
package repository

import (
	"lostfound/internal/model"
	"lostfound/pkg/database"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *model.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByStudentID(studentID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("LOWER(student_id) = LOWER(?)", studentID).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	return &user, err
}

```

### internal/service/auth_service.go

`$lang
package service

import (
	"errors"
	"fmt"
	"lostfound/internal/model"
	"lostfound/internal/repository"
	"lostfound/pkg/utils"
	"strings"
)


type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *AuthService) Register(name, studentID, phone, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	studentID = strings.ToLower(strings.TrimSpace(studentID))
	phone = strings.TrimSpace(phone)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if studentID == "" {
		return nil, errors.New("student ID is required")
	}
	if phone == "" {
		return nil, errors.New("phone number is required")
	}
	if len(strings.TrimSpace(password)) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	existingUser, _ := s.userRepo.FindByStudentID(studentID)
	if existingUser != nil && existingUser.ID > 0 {
		return nil, errors.New("student ID already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:      name,
		StudentID: studentID,
		Phone:     phone,
		Email:     fmt.Sprintf("%s@astu.local", strings.ReplaceAll(strings.ToLower(studentID), "/", "_")),
		Password:  hashedPassword,
		Role:      "student",
	}

	err = s.userRepo.Create(user)
	return user, err
}

func (s *AuthService) Login(studentID, password string) (*model.User, error) {
	studentID = strings.ToLower(strings.TrimSpace(studentID))
	user, err := s.userRepo.FindByStudentID(studentID)
	if err != nil {
		return nil, errors.New("invalid ID or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid ID or password")
	}

	return user, nil
}

```

### internal/service/item_options.go

`$lang
package service

import "strings"

var astuLocations = []string{
	"Library",
	"Cafe",
	"Class",
	"Lab",
	"Dorm",
	"On Road",
	"Toilet",
	"Shower",
	"Amphi",
	"Launch",
	"Park",
	"Hall / Borrow",
	"Other",
}

var colorOptions = []string{
	"red",
	"green",
	"blue",
	"yellow",
	"black",
	"white",
	"gray",
	"brown",
	"orange",
	"purple",
	"pink",
	"gold",
	"silver",
	"other",
}

var itemTypes = []string{
	"lost",
	"found",
}

var itemCategories = []string{
	"electronics",
	"book & document",
	"study tool",
	"atm card",
	"jewelry",
	"sport equipment",
	"bag & backpack",
	"key",
	"id card",
	"clothing & accessories",
	"other",
}

var approvalStatuses = []string{
	"pending",
	"approved",
	"rejected",
}

func ASTULocations() []string {
	return astuLocations
}

func ColorOptions() []string {
	return colorOptions
}

func ItemCategories() []string {
	return itemCategories
}

func IsStandardColor(color string) bool {
	for _, c := range colorOptions {
		if strings.EqualFold(strings.TrimSpace(color), c) {
			return true
		}
	}
	return false
}

func KnownNonOtherColors() []string {
	return []string{"red", "green", "blue", "yellow", "black", "white", "gray", "brown", "orange", "purple", "pink", "gold", "silver"}
}

func IsValidItemType(itemType string) bool {
	for _, t := range itemTypes {
		if strings.EqualFold(strings.TrimSpace(itemType), t) {
			return true
		}
	}
	return false
}

func IsValidCategory(category string) bool {
	for _, c := range itemCategories {
		if strings.EqualFold(strings.TrimSpace(category), c) {
			return true
		}
	}
	return false
}

func IsValidApprovalStatus(status string) bool {
	for _, s := range approvalStatuses {
		if strings.EqualFold(strings.TrimSpace(status), s) {
			return true
		}
	}
	return false
}

func IsValidASTULocation(location string) bool {
	for _, l := range astuLocations {
		if strings.EqualFold(strings.TrimSpace(location), l) {
			return true
		}
	}
	return false
}

```

### internal/service/item_service.go

`$lang
package service

import (
	"errors"
	"fmt"
	"io"
	"lostfound/internal/model"
	"lostfound/internal/repository"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ItemService struct {
	itemRepo *repository.ItemRepository
}

func NewItemService() *ItemService {
	return &ItemService{
		itemRepo: repository.NewItemRepository(),
	}
}

func (s *ItemService) SaveImage(file *multipart.FileHeader) (string, error) {
	uploadDir := "static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	if file.Size > 5*1024*1024 {
		return "", errors.New("image is too large (max 5MB)")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("only JPG and PNG images are allowed")
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the file
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return "/static/uploads/" + filename, nil
}

func (s *ItemService) CreateItem(userID uint, itemType, title, category, color, brand, location, date, description string, imagePaths []string) (*model.Item, error) {
	if !IsValidItemType(itemType) {
		return nil, errors.New("invalid item type")
	}
	if !IsValidCategory(category) {
		return nil, errors.New("invalid category")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("item title is required")
	}
	if strings.TrimSpace(date) != "" {
		if parsed, err := time.Parse("2006-01-02", date); err != nil {
			return nil, errors.New("invalid date format")
		} else if parsed.After(time.Now()) {
			return nil, errors.New("date cannot be in the future")
		}
	}
	primaryImage := ""
	if len(imagePaths) > 0 {
		primaryImage = imagePaths[0]
	}

	item := &model.Item{
		UserID:         userID,
		Type:           itemType,
		Title:          title,
		Category:       category,
		Color:          color,
		Brand:          brand,
		Location:       location,
		Date:           date,
		Description:    description,
		Image:          primaryImage,
		Status:         "open",
		ApprovalStatus: "pending",
	}

	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}
	if err := s.itemRepo.CreateItemImages(item.ID, imagePaths); err != nil {
		return nil, err
	}

	// Notify admins to review the new post (pending).
	_ = s.notifyAdmins(
		"New Post Pending Approval",
		fmt.Sprintf("A %s item was submitted and needs review.\n\nTitle: %s\nCategory: %s\nLocation: %s\nReported by: %d (user id)\n\nOpen Admin > Items to approve/reject.",
			itemType, title, category, location, userID),
	)

	return item, nil
}

func (s *ItemService) SearchItems(filters map[string]interface{}) ([]model.Item, error) {
	return s.itemRepo.FindAll(filters)
}

func (s *ItemService) GetItemByID(id uint) (*model.Item, error) {
	return s.itemRepo.FindByID(id)
}

func (s *ItemService) CreateClaim(itemID, userID uint, description string) error {
	item, err := s.itemRepo.FindByID(itemID)
	if err != nil {
		return err
	}
	if item.UserID == userID {
		return errors.New("you cannot claim your own post")
	}
	if item.ApprovalStatus != "approved" {
		return errors.New("item is not approved by admin yet")
	}
	if item.Status != "open" {
		return errors.New("this item is no longer open for request")
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("request description is required")
	}

	hasActive, err := s.itemRepo.HasActiveClaimByUser(itemID, userID)
	if err != nil {
		return err
	}
	if hasActive {
		return errors.New("you already have a pending or approved request for this post")
	}

	requestType := "claim_request"
	if item.Type == "lost" {
		requestType = "found_match_request"
	}

	claim := &model.Claim{
		ItemID:      itemID,
		UserID:      userID,
		RequestType: requestType,
		Description: description,
		Status:      "pending",
	}

	if err := s.itemRepo.CreateClaim(claim); err != nil {
		return err
	}

	// Notify admins of the incoming request so they don't miss approvals.
	requestTypeLabel := "Claim Request"
	if requestType == "found_match_request" {
		requestTypeLabel = "Found Match Request"
	}
	_ = s.notifyAdmins(
		"New Request Pending",
		fmt.Sprintf(
			"%s submitted for \"%s\".\n\nPost owner: %s (%s / %s)\nRequester: %s (%s / %s)\n\nPlease review in Admin > Claims.",
			requestTypeLabel,
			item.Title,
			item.User.Name, item.User.StudentID, item.User.Phone,
			claim.User.Name, claim.User.StudentID, claim.User.Phone,
		),
	)

	return nil
}

func (s *ItemService) GetStats() (map[string]int64, error) {
	return s.itemRepo.GetStats()
}

func (s *ItemService) GetAllClaims() ([]model.Claim, error) {
	return s.itemRepo.FindAllClaims()
}

func (s *ItemService) UpdateClaimStatus(claimID uint, status, remarks string) error {
	if status != "approved" && status != "rejected" {
		return errors.New("invalid claim status")
	}

	claim, err := s.itemRepo.FindClaimByID(claimID)
	if err != nil {
		return err
	}

	claim.Status = status
	claim.AdminRemarks = remarks

	if status == "approved" {
		item, err := s.itemRepo.FindByID(claim.ItemID)
		if err == nil {
			item.Status = "claimed"
			_ = s.itemRepo.Update(item)
		}

		requestTypeLabel := "Claim Request"
		if claim.RequestType == "found_match_request" {
			requestTypeLabel = "Found Match Request"
		}

		posterName := claim.Item.User.Name
		posterID := claim.Item.User.StudentID
		posterPhone := strings.TrimSpace(claim.Item.User.Phone)
		requesterName := claim.User.Name
		requesterID := claim.User.StudentID
		requesterPhone := strings.TrimSpace(claim.User.Phone)

		requesterBlock := fmt.Sprintf("Contact requester:\nName: %s (%s)\nPhone: %s", requesterName, requesterID, requesterPhone)
		posterBlock := fmt.Sprintf("Contact post owner:\nName: %s (%s)\nPhone: %s", posterName, posterID, posterPhone)
		selfBlockForRequester := fmt.Sprintf("Your details (shared):\nName: %s (%s)\nPhone: %s", requesterName, requesterID, requesterPhone)
		selfBlockForPoster := fmt.Sprintf("Your details (shared):\nName: %s (%s)\nPhone: %s", posterName, posterID, posterPhone)

		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.UserID,
			Title:  "Request Approved",
			Message: fmt.Sprintf(
				"Admin approved your %s for \"%s\".\n\n%s\n\n%s\n\nPlease meet in a safe, public spot on campus and confirm the item details together.",
				requestTypeLabel, claim.Item.Title, posterBlock, selfBlockForRequester,
			),
		})
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.Item.UserID,
			Title:  "Request Approved On Your Post",
			Message: fmt.Sprintf(
				"Admin approved %s on your post \"%s\".\n\n%s\n\n%s\n\nCoordinate handoff safely and mark the item resolved after exchange.",
				requestTypeLabel, claim.Item.Title, requesterBlock, selfBlockForPoster,
			),
		})
	} else {
		reason := strings.TrimSpace(remarks)
		if reason == "" {
			reason = "No remarks provided"
		}
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: claim.UserID,
			Title:  "Request Rejected",
			Message: fmt.Sprintf(
				"Admin rejected your request for \"%s\".\n\nRemarks: %s",
				claim.Item.Title, reason,
			),
		})
	}

	return s.itemRepo.UpdateClaim(claim)
}

func (s *ItemService) GetItemsByUserID(userID uint) ([]model.Item, error) {
	return s.itemRepo.FindByUserID(userID)
}

func (s *ItemService) GetAllItemsForAdmin() ([]model.Item, error) {
	return s.itemRepo.FindAllItemsForAdmin()
}

func (s *ItemService) UpdateItemApproval(itemID uint, status, remarks string) error {
	if !IsValidApprovalStatus(status) {
		return errors.New("invalid approval status")
	}

	item, err := s.itemRepo.FindByID(itemID)
	if err != nil {
		return err
	}

	item.ApprovalStatus = status
	item.AdminRemarks = remarks
	if status != "approved" {
		item.Status = "open"
	}
	if err := s.itemRepo.Update(item); err != nil {
		return err
	}

	if status == "approved" {
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: item.UserID,
			Title:  "Post Approved",
			Message: fmt.Sprintf(
				"Admin approved your %s post \"%s\". It is now visible in the Report list.",
				item.Type, item.Title,
			),
		})
	}
	if status == "rejected" {
		reason := strings.TrimSpace(remarks)
		if reason == "" {
			reason = "No remarks provided"
		}
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: item.UserID,
			Title:  "Post Rejected",
			Message: fmt.Sprintf(
				"Admin rejected your %s post \"%s\". Remarks: %s",
				item.Type, item.Title, reason,
			),
		})
	}

	return nil
}

func (s *ItemService) DeleteItem(itemID uint) error {
	return s.itemRepo.DeleteItem(itemID)
}

func (s *ItemService) HasApprovedRequestForUser(itemID, userID uint) bool {
	ok, err := s.itemRepo.HasApprovedClaimForUser(itemID, userID)
	if err != nil {
		return false
	}
	return ok
}

func (s *ItemService) GetNotificationsByUserID(userID uint) ([]model.Notification, error) {
	return s.itemRepo.FindNotificationsByUserID(userID)
}

func (s *ItemService) MarkNotificationsRead(userID uint) error {
	return s.itemRepo.MarkNotificationsRead(userID)
}

func (s *ItemService) CountUnreadNotifications(userID uint) int64 {
	count, err := s.itemRepo.CountUnreadNotifications(userID)
	if err != nil {
		return 0
	}
	return count
}

func (s *ItemService) notifyAdmins(title, message string) error {
	admins, err := s.itemRepo.FindAdmins()
	if err != nil {
		return err
	}
	for _, admin := range admins {
		_ = s.itemRepo.CreateNotification(&model.Notification{
			UserID: admin.ID,
			Title:  title,
			Message: message,
		})
	}
	return nil
}

```

### pkg/database/db.go

`$lang
package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	var dsn string

	if dbURL != "" {
		dsn = dbURL
	} else {
		host := getEnv("DB_HOST", "localhost")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "")
		name := getEnv("DB_NAME", "lostfound")
		port := getEnv("DB_PORT", "5432")
		sslmode := getEnv("DB_SSLMODE", "disable")

		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, name, port, sslmode,
		)
	}


	// Log (safely hide password)
	logDSN := hidePassword(dsn)
	log.Printf("Connecting with: %s", logDSN)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")
}

// FIX RENDER URL - CORRECTED VERSION
func fixRenderURL(url string) string {
	// 1. Replace postgresql:// with postgres:// for GORM compatibility
	if strings.HasPrefix(url, "postgresql://") {
		url = strings.Replace(url, "postgresql://", "postgres://", 1)
	}

	// 2. Ensure port 5432 is included if not present
	if strings.Contains(url, "postgres://") && !strings.Contains(url, ":5432") {
		// Check if port is missing after @
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			hostPart := parts[1]
			// Split by / to get host:port
			hostPortDb := strings.SplitN(hostPart, "/", 2)
			if len(hostPortDb) == 2 {
				hostPort := hostPortDb[0]
				dbName := hostPortDb[1]
				if !strings.Contains(hostPort, ":") {
					// Add port 5432
					url = parts[0] + "@" + hostPort + ":5432/" + dbName
				}
			}
		}
	}

	// 3. Add sslmode=require if not present (Render PostgreSQL requires SSL)
	if !strings.Contains(url, "sslmode=") {
		if strings.Contains(url, "?") {
			url += "&sslmode=require"
		} else {
			url += "?sslmode=require"
		}
	}

	return url
}

// HIDE PASSWORD IN LOGS - SIMPLIFIED
func hidePassword(dsn string) string {
	if strings.Contains(dsn, "postgres://") {
		return "postgres://****:****@" + strings.SplitN(dsn, "@", 2)[1]
	}
	return dsn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

```

### pkg/utils/hash.go

`$lang
package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

```

### static/css/style.css

`$lang
/* static/css/style.css */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #f5f5f5;
}

.container {
    max-width: 100%;
    margin: 0 auto;
    padding: 0 20px;
}

/* Navbar */
.navbar {
    background: #1a365d;
    color: white;
    padding: 1rem 0;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.navbar .container {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.logo {
    color: white;
    text-decoration: none;
    font-size: 1.5rem;
    font-weight: bold;
}

.nav-links a {
    color: white;
    text-decoration: none;
    margin-left: 1.5rem;
    padding: 0.5rem;
    transition: opacity 0.3s;
}

.nav-links a:hover {
    opacity: 0.8;
}

.notify-link {
    position: relative;
    display: inline-block;
}

.notify-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    margin-left: 2px;
    border-radius: 999px;
    background: #e53e3e;
    color: #fff;
    font-size: 0.65rem;
    font-weight: 700;
    line-height: 1;
    vertical-align: super;
    transform: translateY(-0.35em);
}

.notify-link.unread {
    font-weight: 700;
}

/* Main content */
main {
    min-height: calc(100vh - 140px);
    padding: 2rem 0;
}

/* Footer */
.footer {
    background: #333;
    color: white;
    text-align: center;
    padding: 1rem 0;
    margin-top: auto;
}

/* Buttons */
.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    text-decoration: none;
    font-size: 1rem;
    transition: all 0.3s;
}

.btn-primary {
    background: #1a365d;
    color: white;
}

.btn-primary:hover {
    background: #2d4b7a;
}

.btn-secondary {
    background: #718096;
    color: white;
}

.btn-secondary:hover {
    background: #4a5568;
}

.btn-small {
    padding: 0.25rem 0.75rem;
    font-size: 0.875rem;
    border-radius: 3px;
    background: #1a365d;
    color: white;
    text-decoration: none;
    border: none;
    cursor: pointer;
}

.btn-success {
    background: #38a169;
}

.btn-danger {
    background: #e53e3e;
}

/* Hero section */
.hero {
    text-align: center;
    padding: 4rem 0;
    background: linear-gradient(135deg, #667eea 0%, #1a365d 100%);
    color: white;
    border-radius: 8px;
    margin-bottom: 2rem;
}

.hero h1 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
}

.hero p {
    font-size: 1.2rem;
    margin-bottom: 2rem;
}

.hero-buttons .btn {
    margin: 0 0.5rem;
}

/* Features */
.features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
    margin-top: 2rem;
}

.feature-card {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    text-align: center;
}

.feature-card h3 {
    color: #1a365d;
    margin-bottom: 1rem;
}

/* Forms */
.auth-form, .report-form, .search-page {
    max-width: 500px;
    margin: 0 auto;
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
}

.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

.error {
    background: #fee;
    color: #c33;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    border: 1px solid #fcc;
}

/* Stats */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1.5rem;
    margin: 2rem 0;
}

.stat-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    text-align: center;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.stat-number {
    font-size: 2.5rem;
    font-weight: bold;
    color: #1a365d;
}

/* Items grid */
.items-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-top: 2rem;
}

.item-card {
    background: white;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.item-image {
    width: 100%;
    height: 200px;
    object-fit: cover;
}

.no-image {
    height: 200px;
    background: #f0f0f0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
}

.item-details {
    padding: 1rem;
}

/* Item detail */
.item-detail {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.item-content {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
    margin-bottom: 2rem;
}

.item-image-large img {
    width: 100%;
    max-height: 400px;
    object-fit: contain;
    cursor: zoom-in;
}

.item-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    margin-top: 1.5rem;
}

.item-actions .btn {
    min-width: 170px;
    text-align: center;
}

.item-gallery {
    display: flex;
    gap: 8px;
    margin-top: 10px;
    flex-wrap: wrap;
}

.item-thumb {
    width: 90px;
    height: 90px;
    object-fit: cover;
    border-radius: 6px;
    border: 1px solid #ddd;
    cursor: pointer;
}

.image-modal {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    z-index: 9999;
    display: none;
    align-items: center;
    justify-content: center;
    padding: 20px;
}

.image-modal-content {
    max-width: 95vw;
    max-height: 90vh;
    object-fit: contain;
    border-radius: 8px;
}

.image-modal-close {
    position: absolute;
    top: 18px;
    right: 24px;
    font-size: 2rem;
    color: #fff;
    cursor: pointer;
}

.image-nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    border: none;
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.2);
    color: #fff;
    font-size: 1.5rem;
    cursor: pointer;
}

.image-nav-prev {
    left: 18px;
}

.image-nav-next {
    right: 18px;
}

.image-modal-counter {
    position: absolute;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    color: #fff;
    background: rgba(0, 0, 0, 0.5);
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 0.85rem;
}

/* Badges */
.badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 3px;
    font-size: 0.875rem;
    font-weight: 500;
}

.badge.lost {
    background: #feb2b2;
    color: #742a2a;
}

.badge.found {
    background: #9ae6b4;
    color: #22543d;
}

.badge.pending {
    background: #fbd38d;
    color: #744210;
}

.badge.approved {
    background: #9ae6b4;
    color: #22543d;
}

.badge.rejected {
    background: #feb2b2;
    color: #742a2a;
}

/* Admin tables */
.admin-table {
    width: 100%;
    background: white;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    margin-top: 1rem;
}

.admin-table th,
.admin-table td {
    padding: 1rem;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

.admin-table th {
    background: #f7fafc;
    font-weight: 600;
}

.inline-form {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.admin-filter {
    max-width: 100%;
    margin-bottom: 1rem;
    padding: 1rem;
}

.admin-image-grid {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
}

.admin-item-thumb {
    width: 56px;
    height: 56px;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid #ddd;
}

/* Dashboard actions */
.dashboard-actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
}

/* Auth link */
.auth-link {
    text-align: center;
    margin-top: 1rem;
}

.demo-credentials {
    margin-top: 2rem;
    padding: 1rem;
    background: #f7fafc;
    border-radius: 4px;
    font-size: 0.875rem;
}

.notifications-table .notif-unread td {
    background: #fff8db;
}

.file-input-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
}

.upload-widget {
    position: relative;
}

.upload-dropzone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 130px;
    border: 2px dashed #7a8ca5;
    border-radius: 8px;
    background: #f8fbff;
    color: #243b5a;
    cursor: pointer;
    padding: 16px;
    text-align: center;
}

.upload-dropzone.drag-active {
    border-color: #1a365d;
    background: #e8f0ff;
}

.upload-actions {
    display: flex;
    gap: 10px;
    margin-top: 10px;
}

.upload-preview {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
    gap: 10px;
    margin-top: 12px;
}

.upload-preview-item {
    text-align: center;
}

.upload-preview-item img {
    width: 100%;
    height: 90px;
    object-fit: cover;
    border-radius: 6px;
    border: 1px solid #ddd;
}

.upload-preview-item small {
    display: block;
    margin-top: 4px;
    color: #555;
    word-break: break-word;
}

/* Responsive */
@media (max-width: 768px) {
    .hero h1 {
        font-size: 2rem;
    }
    
    .item-content {
        grid-template-columns: 1fr;
    }
    
    .form-row {
        grid-template-columns: 1fr;
    }
    
    .dashboard-actions {
        flex-direction: column;
    }

    .item-actions {
        flex-direction: column;
    }

    .upload-actions {
        flex-direction: column;
    }

    .image-nav {
        width: 38px;
        height: 38px;
    }
    
    .btn {
        width: 100%;
        text-align: center;
    }
}
/* Upload widget overrides */
.upload-widget { position: relative; }
.upload-dropzone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 150px;
    border: 2px dashed #94a3b8;
    border-radius: 12px;
    background: #f8fafc;
    color: #1f2937;
    cursor: pointer;
    padding: 18px;
    text-align: center;
    transition: border-color 0.2s ease, background 0.2s ease;
}
.upload-dropzone.drag-active {
    border-color: #0f172a;
    background: #e2e8f0;
}
.file-input-hidden { 
    position: absolute;
    left: -9999px;
    top: auto;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
    overflow: hidden;
}
.upload-preview { margin-top: 10px; display: flex; flex-wrap: wrap; gap: 8px; }
.upload-preview .badge { background: #e2e8f0; color: #0f172a; border-radius: 999px; padding: 4px 8px; font-size: 12px; }

/* Card slider + lightbox */
.slider-nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 40px;
    height: 40px;
    border: none;
    border-radius: 999px;
    background: rgba(15, 23, 42, 0.65);
    color: #fff;
    display: none;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    cursor: pointer;
    transition: background 0.2s ease;
}
.slider-nav:hover { background: rgba(15, 23, 42, 0.85); }
.slider-prev { left: 8px; }
.slider-next { right: 8px; }
.slider-dots {
    position: absolute;
    bottom: 6px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: 6px;
}
.slider-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: rgba(15, 23, 42, 0.25);
    border: 1px solid rgba(255,255,255,0.4);
    cursor: pointer;
}
.slider-dot.active { background: #0f172a; border-color: #fff; }

@media (hover: none) {
    .slider-nav {
        display: flex;
    }
}

.lightbox-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.92);
    z-index: 9999;
    display: none;
    align-items: center;
    justify-content: center;
    padding: 24px;
}
.lightbox-content {
    position: relative;
    max-width: 95vw;
    max-height: 90vh;
}
.lightbox-image {
    max-width: 95vw;
    max-height: 90vh;
    object-fit: contain;
    border-radius: 10px;
    box-shadow: 0 10px 40px rgba(0,0,0,0.35);
}
.lightbox-close {
    position: absolute;
    top: -14px;
    right: -14px;
    background: rgba(15, 23, 42, 0.9);
    color: #fff;
    border: none;
    width: 34px;
    height: 34px;
    border-radius: 999px;
    font-size: 18px;
    cursor: pointer;
    box-shadow: 0 4px 10px rgba(0,0,0,0.35);
}
.lightbox-nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    background: rgba(15,23,42,0.8);
    color: #fff;
    border: none;
    width: 46px;
    height: 46px;
    border-radius: 999px;
    font-size: 20px;
    cursor: pointer;
}
.lightbox-prev { left: -58px; }
.lightbox-next { right: -58px; }

@media (max-width: 640px) {
    .slider-nav { width: 36px; height: 36px; display: flex; }
    .slider-prev { left: 6px; }
    .slider-next { right: 6px; }
    .slider-dot { width: 10px; height: 10px; }
    .lightbox-overlay { padding: 12px; }
    .lightbox-nav { width: 42px; height: 42px; }
    .lightbox-prev { left: -48px; }
    .lightbox-next { right: -48px; }
}

```

### static/js/image-upload.js

`$lang
// Drag and drop functionality for image uploads
document.addEventListener('DOMContentLoaded', function() {
    const dropZone = document.querySelector('.upload-dropzone');
    const fileInput = document.getElementById('images');
    const uploadPreview = document.querySelector('.upload-preview');
    const MAX_FILES = 5;

    if (!dropZone || !fileInput) return;

    let selection = new DataTransfer();

    dropZone.addEventListener('click', () => fileInput.click());

    // Prevent default drag behaviors
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
        document.body.addEventListener(eventName, preventDefaults, false);
    });

    // Highlight drop area when item is dragged over it
    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, unhighlight, false);
    });

    // Handle dropped files
    dropZone.addEventListener('drop', handleDrop, false);

    // Handle file selection via input
    fileInput.addEventListener('change', (e) => handleFiles(e, true), false);

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    function highlight() {
        dropZone.classList.add('drag-active');
    }

    function unhighlight() {
        dropZone.classList.remove('drag-active');
    }

    function handleDrop(e) {
        const dt = e.dataTransfer;
        handleFiles({ target: { files: dt.files } }, true);
    }

    function handleFiles(e, reset = false) {
        const files = e.target.files;
        if (!files || files.length === 0) return;

        if (reset) {
            selection = new DataTransfer();
        }

        for (let i = 0; i < files.length; i++) {
            const file = files[i];

            if (!file.type.match('image.*')) {
                alert('Please select only image files (JPG, PNG)');
                continue;
            }

            if (file.size > 5 * 1024 * 1024) {
                alert(`File ${file.name} is too large (max 5MB)`);
                continue;
            }

            if (selection.files.length >= MAX_FILES) {
                alert(`You can upload up to ${MAX_FILES} images.`);
                break;
            }

            selection.items.add(file);
        }

        syncSelection();
    }

    function syncSelection() {
        fileInput.files = selection.files;
        renderPreview();
    }

    function renderPreview() {
        if (!uploadPreview) return;
        uploadPreview.innerHTML = '';

        Array.from(selection.files).forEach(file => {
            const reader = new FileReader();
            reader.onload = function(e) {
                const previewContainer = document.createElement('div');
                previewContainer.className = 'upload-preview-item';

                const img = document.createElement('img');
                img.src = e.target.result;
                img.alt = file.name;

                const fileName = document.createElement('small');
                fileName.textContent = truncateFileName(file.name, 20);

                previewContainer.appendChild(img);
                previewContainer.appendChild(fileName);

                uploadPreview.appendChild(previewContainer);
            };
            reader.readAsDataURL(file);
        });
    }

    function truncateFileName(name, maxLength) {
        if (name.length <= maxLength) return name;
        return name.substr(0, maxLength) + '...';
    }
});

```

### templates/layout.html

`$lang
<!-- templates/layout.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ if .title }}{{ .title }} - ASTU Lost & Found{{ else }}ASTU Lost & Found{{ end }}</title>
    <link rel="icon" href="/static/img/logo.jpg.png" type="image/x-icon">
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="/static/js/image-upload.js"></script>
</head>
<body class="min-h-screen bg-slate-100 text-slate-900">
    <nav class="w-full bg-slate-950 text-white border-b border-slate-800 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex items-center h-16 gap-3">
                <a href="/" class="flex items-center gap-2 shrink-0 pr-3 md:pr-4 border-r border-slate-800">
                    <div class="h-8 w-8 rounded-md bg-slate-900 flex items-center justify-center">
                        <span class="text-white font-bold text-sm">LF</span>
                    </div>
                    <span class="text-sm md:text-base font-bold text-white">ASTU L&amp;F</span>
                </a>

                <button class="ml-1 inline-flex items-center justify-center h-9 w-9 rounded-md bg-slate-900 text-white md:hidden" id="nav-toggle" aria-label="Toggle menu">
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                </button>

                <div class="hidden md:flex items-center gap-1 flex-wrap" id="nav-links">
                    <a href="/" class="px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Home</a>
                    <a href="/report" class="px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report View</a>
                    {{ if .user }}
                        <a href="/report/new?type=lost" class="px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report Lost</a>
                        <a href="/report/new?type=found" class="px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report Found</a>
                        <a href="/notifications" class="relative px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition {{ if gt .unread_count 0 }}font-bold{{ end }}">
                            Notifications
                            {{ if gt .unread_count 0 }}
                              <span class="absolute top-1 right-1 block h-2 w-2 rounded-full bg-rose-500"></span>
                            {{ end }}
                        </a>
                    {{ end }}
                </div>

                <div class="ml-auto flex items-center gap-3 pl-3 border-l border-slate-800">
                    {{ if .user }}
                        <span class="text-xs md:text-sm font-semibold text-slate-200 truncate max-w-[120px]">{{ .user.Name }}</span>
                        <a href="/logout" class="text-xs font-semibold text-slate-200 hover:text-white">Logout</a>
                    {{ else }}
                        <a href="/login" class="text-xs font-semibold text-slate-200 hover:text-white">Login</a>
                        <a href="/register" class="text-xs font-semibold text-slate-200 hover:text-white">Register</a>
                    {{ end }}
                </div>
            </div>

            <div class="md:hidden hidden flex-col gap-1 pb-3" id="nav-links-mobile">
                <a href="/" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Home</a>
                <a href="/report" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report View</a>
                {{ if .user }}
                    <a href="/report/new?type=lost" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report Lost</a>
                    <a href="/report/new?type=found" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Report Found</a>
                    <a href="/notifications" class="relative block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition {{ if gt .unread_count 0 }}font-bold{{ end }}">
                        Notifications
                        {{ if gt .unread_count 0 }}
                          <span class="absolute top-2 right-3 block h-2 w-2 rounded-full bg-rose-500"></span>
                        {{ end }}
                    </a>
                {{ else }}
                    <a href="/login" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Login</a>
                    <a href="/register" class="block px-3 py-2 rounded-md text-xs font-semibold text-slate-200 hover:text-white hover:bg-slate-900 transition">Register</a>
                {{ end }}
            </div>
        </div>
    </nav>

    <script>
    (() => {
      const toggle = document.getElementById('nav-toggle');
      const mobile = document.getElementById('nav-links-mobile');
      if (!toggle || !mobile) return;
      toggle.addEventListener('click', () => {
        mobile.classList.toggle('hidden');
      });
    })();
    </script>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {{ if eq .content_template "index_content" }}
            {{ template "index_content" . }}
        {{ else if eq .content_template "login_content" }}
            {{ template "login_content" . }}
        {{ else if eq .content_template "register_content" }}
            {{ template "register_content" . }}
        {{ else if eq .content_template "dashboard_content" }}
            {{ template "dashboard_content" . }}
        {{ else if eq .content_template "report_content" }}
            {{ template "report_content" . }}
        {{ else if eq .content_template "items_content" }}
            {{ template "items_content" . }}
        {{ else if eq .content_template "item_content" }}
            {{ template "item_content" . }}
        {{ else if eq .content_template "notifications_content" }}
            {{ template "notifications_content" . }}
        {{ else if eq .content_template "admin_dashboard_content" }}
            {{ template "admin_dashboard_content" . }}
        {{ else if eq .content_template "admin_claims_content" }}
            {{ template "admin_claims_content" . }}
        {{ else if eq .content_template "admin_items_content" }}
            {{ template "admin_items_content" . }}
        {{ else if eq .content_template "error_content" }}
            {{ template "error_content" . }}
        {{ end }}
    </main>

    <footer class="w-full bg-slate-900 text-white">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 text-sm">
            &copy; 2024 ASTU Lost &amp; Found System
        </div>
    </footer>
</body>
</html>

```

### templates/index.html

`$lang
<!-- templates/index.html -->
{{ template "layout.html" . }}

{{ define "index_content" }}
<section class="bg-white rounded-2xl border border-slate-200 shadow-sm px-6 py-10 md:px-10 md:py-12">
  <div class="max-w-4xl space-y-5">
    <p class="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">ASTU Digital Lost &amp; Found</p>
    <h1 class="text-3xl md:text-4xl font-extrabold text-slate-900 leading-snug">
      Reunite the ASTU community with lost and found items.
    </h1>
    <p class="text-base text-slate-600 md:text-lg leading-relaxed">
      Report what you’ve lost or found, let admins verify, and connect securely with the right person when a match is approved.
    </p>
    <div class="flex flex-wrap gap-3 pt-2">
      <a href="/report/new?type=lost" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-slate-900 text-white text-sm font-semibold shadow hover:bg-slate-800 transition">
        Report Lost
      </a>
      <a href="/report/new?type=found" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-emerald-600 text-white text-sm font-semibold shadow hover:bg-emerald-500 transition">
        Report Found
      </a>
      <a href="/report" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-slate-100 text-slate-800 text-sm font-semibold hover:bg-slate-200 transition">
        Browse Reports
      </a>
    </div>
  </div>
</section>

<section class="mt-10 grid md:grid-cols-3 gap-4">
  <div class="bg-white rounded-xl p-5 shadow-sm border border-slate-200">
    <h3 class="text-lg font-semibold text-slate-900">Simple reporting</h3>
    <p class="text-slate-600 text-sm mt-2">Lost or found? Submit details quickly with photos and campus location.</p>
  </div>
  <div class="bg-white rounded-xl p-5 shadow-sm border border-slate-200">
    <h3 class="text-lg font-semibold text-slate-900">Admin verification</h3>
    <p class="text-slate-600 text-sm mt-2">Admins approve listings so only verified posts appear to everyone.</p>
  </div>
  <div class="bg-white rounded-xl p-5 shadow-sm border border-slate-200">
    <h3 class="text-lg font-semibold text-slate-900">Safe connections</h3>
    <p class="text-slate-600 text-sm mt-2">When a claim is approved, both sides get contact info to meet offline.</p>
  </div>
</section>
{{ end }}

```

### templates/items.html

`$lang
<!-- templates/items.html -->
{{ template "layout.html" . }}

{{ define "items_content" }}
<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
    <div>
      <h2 class="text-3xl font-bold text-slate-900">Find Your Items</h2>
      <p class="mt-2 text-slate-600">Browse all reported lost and found items</p>
    </div>
    {{ if .user }}
    <div class="flex flex-wrap gap-3">
      <a href="/report/new?type=lost" class="flex-1 min-w-[140px] px-4 py-3 bg-gradient-to-r from-rose-600 to-rose-700 text-white rounded-xl text-center font-semibold shadow-lg hover:shadow-xl transition-all duration-300 transform hover:-translate-y-0.5">Report Lost</a>
      <a href="/report/new?type=found" class="flex-1 min-w-[140px] px-4 py-3 bg-gradient-to-r from-emerald-600 to-emerald-700 text-white rounded-xl text-center font-semibold shadow-lg hover:shadow-xl transition-all duration-300 transform hover:-translate-y-0.5">Report Found</a>
    </div>
    {{ end }}
  </div>

  <div class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="p-5 border-b border-slate-200">
      <button id="filter-toggle" type="button" class="w-full flex items-center justify-between">
        <span class="font-semibold text-slate-800">Filters</span>
        <svg class="h-5 w-5 text-slate-500 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
    </div>
    <div id="filter-panel" class="p-5">
      <form method="GET" action="/report" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Type</label>
          <select name="type" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
            <option value="">All Types</option>
            <option value="lost" {{ if eq .selected_type "lost" }}selected{{ end }}>Lost</option>
            <option value="found" {{ if eq .selected_type "found" }}selected{{ end }}>Found</option>
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Category</label>
          <select name="category" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
            <option value="">All Categories</option>
            {{ range .categories }}
            <option value="{{ . }}" {{ if eq $.selected_category . }}selected{{ end }}>{{ . }}</option>
            {{ end }}
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Location</label>
          <select name="location" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
            <option value="">All Locations</option>
            {{ range .locations }}
            <option value="{{ . }}" {{ if eq $.selected_location . }}selected{{ end }}>{{ . }}</option>
            {{ end }}
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Color</label>
          <select name="color" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
            <option value="">All Colors</option>
            {{ range .colors }}
            <option value="{{ . }}" {{ if eq $.selected_color . }}selected{{ end }}>{{ . }}</option>
            {{ end }}
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Date</label>
          <input type="date" name="date_from" value="{{ .selected_date_from }}" max="{{ now }}" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>
        <div class="md:col-span-2 lg:col-span-5 flex gap-3 pt-2">
          <button type="submit" class="px-5 py-2.5 bg-slate-900 text-white rounded-lg font-medium hover:bg-slate-800 transition flex items-center">
            <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            Apply Filters
          </button>
          <a href="/report" class="px-5 py-2.5 bg-slate-100 text-slate-700 rounded-lg font-medium hover:bg-slate-200 transition flex items-center">Clear All</a>
        </div>
      </form>
    </div>
  </div>

  {{ if .items }}
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    {{ range .items }}
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden hover:shadow-lg transition-all duration-300 transform hover:-translate-y-1">
      <div class="relative block bg-slate-50">
        {{ if .Images }}
          {{ $count := len .Images }}
          <div class="relative h-52 bg-slate-50 rounded-md overflow-hidden group cursor-pointer" data-gallery data-count="{{ $count }}">
            {{ range $i, $img := .Images }}
              <img src="{{ $img.Path }}" alt="{{ $.Title }}" class="slider-frame w-full h-full object-contain p-2 {{ if eq $i 0 }}block{{ else }}hidden{{ end }}" data-idx="{{ $i }}">
            {{ end }}
            {{ if gt $count 1 }}
              <button type="button" class="slider-nav slider-prev hidden group-hover:flex" aria-label="Previous image">&#10094;</button>
              <button type="button" class="slider-nav slider-next hidden group-hover:flex" aria-label="Next image">&#10095;</button>
              <div class="slider-dots"></div>
            {{ end }}
          </div>
        {{ else if .Image }}
          <div class="relative h-52 bg-slate-50 rounded-md overflow-hidden cursor-pointer" data-gallery data-count="1">
            <img src="{{ .Image }}" alt="{{ .Title }}" class="w-full h-full object-contain p-2" data-idx="0">
          </div>
        {{ else }}
          <div class="h-52 flex items-center justify-center bg-gradient-to-br from-slate-100 to-slate-200">
            <svg class="h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        {{ end }}
        <div class="absolute top-3 right-3">
          <span class="px-3 py-1 rounded-full text-xs font-semibold {{ if eq .Type "lost" }}bg-rose-100 text-rose-800{{ else }}bg-emerald-100 text-emerald-800{{ end }}">
            {{ if eq .Type "lost" }}Lost{{ else }}Found{{ end }}
          </span>
        </div>
      </div>
      <div class="p-5">
        <h3 class="text-lg font-bold text-slate-900 mb-2">{{ .Title }}</h3>
        <div class="space-y-2 text-sm text-slate-600">
          <div class="flex items-center">
            <svg class="h-4 w-4 mr-2 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
            </svg>
            <span>{{ .Category }}</span>
          </div>
          <div class="flex items-center">
            <svg class="h-4 w-4 mr-2 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span>
              {{ if and (eq .Type "found") (not $.is_admin) }}
                Hidden until approved
              {{ else }}
                {{ .Location }}
              {{ end }}
            </span>
          </div>
          <div class="flex items-center">
            <svg class="h-4 w-4 mr-2 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span>{{ .Date }}</span>
          </div>
        </div>
        <div class="mt-4">
          <a href="/item/{{ .ID }}" class="w-full flex items-center justify-center px-4 py-2.5 bg-slate-900 text-white rounded-lg font-medium hover:bg-slate-800 transition">View Details</a>
        </div>
      </div>
    </div>
    {{ end }}
  </div>
  {{ else }}
  <div class="bg-white rounded-2xl shadow-sm border border-slate-200 p-12 text-center">
    <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
    <h3 class="mt-4 text-lg font-medium text-slate-900">No items found</h3>
    <p class="mt-2 text-slate-500">Try adjusting your search filters</p>
  </div>
  {{ end }}
</div>

<script>
(() => {
  const toggle = document.getElementById('filter-toggle');
  const panel = document.getElementById('filter-panel');
  const icon = toggle ? toggle.querySelector('svg') : null;
  if (!toggle || !panel) return;
  toggle.addEventListener('click', () => {
    const hidden = panel.classList.toggle('hidden');
    if (icon) icon.style.transform = hidden ? 'rotate(0deg)' : 'rotate(180deg)';
  });
})();

// Lightweight slider + lightbox for item cards
(() => {
  const galleries = document.querySelectorAll('[data-gallery]');
  if (!galleries.length) return;

  // Build lightbox once
  const overlay = document.createElement('div');
  overlay.className = 'lightbox-overlay';
  overlay.innerHTML = `
    <div class="lightbox-content">
      <button class="lightbox-close" aria-label="Close">&times;</button>
      <button class="lightbox-nav lightbox-prev" aria-label="Previous">&#10094;</button>
      <img class="lightbox-image" src="" alt="Preview">
      <button class="lightbox-nav lightbox-next" aria-label="Next">&#10095;</button>
    </div>
  `;
  document.body.appendChild(overlay);
  const lbImg = overlay.querySelector('.lightbox-image');
  const lbClose = overlay.querySelector('.lightbox-close');
  const lbPrev = overlay.querySelector('.lightbox-prev');
  const lbNext = overlay.querySelector('.lightbox-next');

  let currentList = [];
  let currentIndex = 0;

  function openLightbox(list, idx) {
    currentList = list;
    currentIndex = idx;
    lbImg.src = currentList[currentIndex];
    overlay.style.display = 'flex';
  }
  function closeLightbox() {
    overlay.style.display = 'none';
    lbImg.src = '';
  }
  function shift(delta) {
    if (!currentList.length) return;
    currentIndex = (currentIndex + delta + currentList.length) % currentList.length;
    lbImg.src = currentList[currentIndex];
  }

  lbClose.addEventListener('click', closeLightbox);
  overlay.addEventListener('click', (e) => { if (e.target === overlay) closeLightbox(); });
  lbPrev.addEventListener('click', () => shift(-1));
  lbNext.addEventListener('click', () => shift(1));
  document.addEventListener('keydown', (e) => {
    if (overlay.style.display !== 'flex') return;
    if (e.key === 'Escape') closeLightbox();
    if (e.key === 'ArrowLeft') shift(-1);
    if (e.key === 'ArrowRight') shift(1);
  });

  galleries.forEach(gal => {
    const frames = gal.querySelectorAll('.slider-frame, img[data-idx]');
    if (!frames.length) return;
    const dotsHost = gal.querySelector('.slider-dots');
    let active = 0;
    const hasMulti = frames.length > 1;

    function setActive(i) {
      active = i;
      frames.forEach((f, idx) => f.classList.toggle('hidden', idx !== active));
      if (dotsHost) {
        dotsHost.querySelectorAll('.slider-dot').forEach((d, idx) => d.classList.toggle('active', idx === active));
      }
    }

    // dots
    if (dotsHost && hasMulti) {
      dotsHost.innerHTML = '';
      frames.forEach((_, idx) => {
        const dot = document.createElement('button');
        dot.type = 'button';
        dot.className = 'slider-dot' + (idx === 0 ? ' active' : '');
        dot.addEventListener('click', (e) => { e.stopPropagation(); setActive(idx); });
        dotsHost.appendChild(dot);
      });
    }

    const prev = gal.querySelector('.slider-prev');
    const next = gal.querySelector('.slider-next');
    if (prev && next && hasMulti) {
      prev.addEventListener('click', (e) => { e.stopPropagation(); setActive((active - 1 + frames.length) % frames.length); });
      next.addEventListener('click', (e) => { e.stopPropagation(); setActive((active + 1) % frames.length); });
    }

    gal.addEventListener('click', () => {
      const list = Array.from(frames).map(f => f.getAttribute('src'));
      openLightbox(list, active);
    });
  });
})();
</script>
{{ end }}

```

### templates/item.html

`$lang
<!-- templates/item.html -->
{{ template "layout.html" . }}

{{ define "item_content" }}
<div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
  <div class="grid lg:grid-cols-2 gap-0">
    <div class="p-6 bg-slate-50">
      {{ if or .item.Images .item.Image }}
        {{ if .item.Images }}
          {{ $count := len .item.Images }}
          <div class="relative h-80 bg-white rounded-xl shadow-sm p-2 flex items-center justify-center cursor-pointer" data-item-gallery data-count="{{ $count }}">
            {{ range $i, $img := .item.Images }}
              <img src="{{ $img.Path }}" alt="{{ $.item.Title }}" class="item-frame max-h-full max-w-full object-contain {{ if eq $i 0 }}block{{ else }}hidden{{ end }}" data-idx="{{ $i }}">
            {{ end }}
            {{ if gt $count 1 }}
              <button type="button" class="slider-nav slider-prev" aria-label="Previous image">&#10094;</button>
              <button type="button" class="slider-nav slider-next" aria-label="Next image">&#10095;</button>
            {{ end }}
          </div>
          <div class="flex flex-wrap gap-2 mt-3">
            {{ range $i, $img := .item.Images }}
            <button type="button" class="rounded-lg border border-slate-200 hover:border-slate-400 transition" data-thumb-idx="{{ $i }}">
              <img src="{{ $img.Path }}" alt="item image" class="w-20 h-20 object-contain bg-white rounded-md p-1">
            </button>
            {{ end }}
          </div>
        {{ else }}
          <div class="relative h-80 bg-white rounded-xl shadow-sm p-2 flex items-center justify-center cursor-pointer" data-item-gallery data-count="1">
            <img src="{{ .item.Image }}" alt="{{ .item.Title }}" class="max-h-full max-w-full object-contain" data-idx="0">
          </div>
        {{ end }}
      {{ else }}
        <div class="w-full h-80 bg-slate-100 rounded-xl flex items-center justify-center text-slate-400">No Image</div>
      {{ end }}
    </div>

    <div class="p-6 space-y-3">
      <div class="flex items-center gap-2">
        <h2 class="text-2xl font-bold text-slate-900 flex-1">{{ .item.Title }}</h2>
        <span class="px-3 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800 capitalize">{{ .item.Type }}</span>
      </div>
      <p class="text-sm text-slate-600"><span class="font-semibold">Category:</span> {{ .item.Category }}</p>
      <p class="text-sm text-slate-600"><span class="font-semibold">Color:</span> {{ .item.Color }}</p>
      <p class="text-sm text-slate-600"><span class="font-semibold">Brand:</span> {{ .item.Brand }}</p>
      <p class="text-sm text-slate-600">
        <span class="font-semibold">Location:</span>
        {{ if and (eq .item.Type "found") (not .show_private) }}
          Hidden until admin approves your request
        {{ else }}
          {{ .item.Location }}
        {{ end }}
      </p>
      <p class="text-sm text-slate-600"><span class="font-semibold">Date:</span> {{ .item.Date }}</p>
      <p class="text-sm text-slate-600">
        <span class="font-semibold">Reported by:</span>
        {{ if .show_private }}{{ .item.User.Name }} ({{ .item.User.StudentID }}){{ else }}Hidden until approval{{ end }}
      </p>
      <p class="text-sm text-slate-600">
        <span class="font-semibold">Contact:</span>
        {{ if .show_private }}{{ .item.User.Phone }}{{ else }}Available after admin approval{{ end }}
      </p>
      <p class="text-sm text-slate-600"><span class="font-semibold">Approval:</span> {{ .item.ApprovalStatus }}</p>
      <div class="text-sm text-slate-700 leading-relaxed">
        <p class="font-semibold text-slate-900 mb-1">Description</p>
        <p>{{ .item.Description }}</p>
      </div>

      {{ if .can_request }}
      <div class="mt-4 border border-slate-200 rounded-xl p-4 bg-slate-50">
        <h3 class="text-sm font-semibold text-slate-900 mb-2">{{ .request_type }}</h3>
        <form method="POST" action="/claim" class="space-y-3">
          <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
          <input type="hidden" name="item_id" value="{{ .item.ID }}">
          <div class="grid md:grid-cols-2 gap-3">
            <div>
              <label class="text-xs font-semibold text-slate-600">Location (where you found/saw it)</label>
              <select name="claim_location" class="w-full px-3 py-2 mt-1 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
                <option value="">Select location</option>
                {{ range .locations }}<option value="{{ . }}">{{ . }}</option>{{ end }}
              </select>
            </div>
            <div>
              <label class="text-xs font-semibold text-slate-600">Category</label>
              <select name="claim_category" class="w-full px-3 py-2 mt-1 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
                <option value="">Select category</option>
                {{ range .categories }}<option value="{{ . }}">{{ . }}</option>{{ end }}
              </select>
            </div>
            <div>
              <label class="text-xs font-semibold text-slate-600">Color</label>
              <select name="claim_color" class="w-full px-3 py-2 mt-1 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
                <option value="">Select color</option>
                {{ range .colors }}<option value="{{ . }}">{{ . }}</option>{{ end }}
              </select>
            </div>
            <div>
              <label class="text-xs font-semibold text-slate-600">Date</label>
              <input type="date" name="claim_date" max="{{ now }}" class="w-full px-3 py-2 mt-1 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
            </div>
          </div>
          <div>
            <label class="text-xs font-semibold text-slate-600">{{ .request_hint }}</label>
            <textarea name="description" rows="3" class="w-full px-3 py-2 mt-1 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent" placeholder="Add proof or extra details..."></textarea>
          </div>
          <button type="submit" class="bg-slate-900 text-white px-4 py-2 rounded-lg text-sm font-semibold shadow hover:-translate-y-[1px] transition">Submit Request</button>
        </form>
      </div>
      {{ end }}

      {{ if and (not .user) (eq .item.ApprovalStatus "approved") (eq .item.Status "open") }}
      <div class="mt-3 p-3 border border-dashed border-slate-300 rounded-lg text-sm text-slate-700">
        Want to request this item? <a href="/login" class="font-semibold text-sky-700 hover:text-sky-900">Login first</a>.
      </div>
      {{ end }}

      <div class="pt-4 flex gap-2">
        <a href="/report" class="bg-white border border-slate-300 text-slate-800 px-4 py-2 rounded-lg text-sm font-semibold hover:-translate-y-[1px] transition">Back to Reports</a>
      </div>
    </div>
  </div>
</div>
{{ end }}

<script>
(() => {
  const gallery = document.querySelector('[data-item-gallery]');
  if (!gallery) return;

  const frames = gallery.querySelectorAll('.item-frame, img[data-idx]');
  const thumbs = document.querySelectorAll('[data-thumb-idx]');
  const prev = gallery.querySelector('.slider-prev');
  const next = gallery.querySelector('.slider-next');
  const hasMulti = frames.length > 1;
  let active = 0;

  const overlay = document.createElement('div');
  overlay.className = 'lightbox-overlay';
  overlay.innerHTML = `
    <div class="lightbox-content">
      <button class="lightbox-close" aria-label="Close">&times;</button>
      ${hasMulti ? '<button class="lightbox-nav lightbox-prev" aria-label="Previous">&#10094;</button>' : ''}
      <img class="lightbox-image" src="" alt="Preview">
      ${hasMulti ? '<button class="lightbox-nav lightbox-next" aria-label="Next">&#10095;</button>' : ''}
    </div>
  `;
  document.body.appendChild(overlay);
  const lbImg = overlay.querySelector('.lightbox-image');
  const lbClose = overlay.querySelector('.lightbox-close');
  const lbPrev = overlay.querySelector('.lightbox-prev');
  const lbNext = overlay.querySelector('.lightbox-next');

  function setActive(idx) {
    active = idx;
    frames.forEach((f, i) => f.classList.toggle('hidden', i !== active));
    thumbs.forEach((t, i) => t.classList.toggle('ring-2', i === active));
  }

  function openLightbox() {
    lbImg.src = frames[active].getAttribute('src');
    overlay.style.display = 'flex';
  }
  function closeLightbox() {
    overlay.style.display = 'none';
  }
  function shift(delta) {
    if (!hasMulti) return;
    active = (active + delta + frames.length) % frames.length;
    setActive(active);
    lbImg.src = frames[active].getAttribute('src');
  }

  gallery.addEventListener('click', () => {
    setActive(active);
    openLightbox();
  });
  if (prev && next && hasMulti) {
    prev.addEventListener('click', (e) => { e.stopPropagation(); shift(-1); });
    next.addEventListener('click', (e) => { e.stopPropagation(); shift(1); });
  }
  thumbs.forEach((t, i) => {
    t.addEventListener('click', (e) => {
      e.stopPropagation();
      setActive(i);
    });
  });
  lbClose.addEventListener('click', closeLightbox);
  overlay.addEventListener('click', (e) => { if (e.target === overlay) closeLightbox(); });
  if (lbPrev) lbPrev.addEventListener('click', (e) => { e.stopPropagation(); shift(-1); });
  if (lbNext) lbNext.addEventListener('click', (e) => { e.stopPropagation(); shift(1); });
  document.addEventListener('keydown', (e) => {
    if (overlay.style.display !== 'flex') return;
    if (e.key === 'Escape') closeLightbox();
    if (e.key === 'ArrowLeft') shift(-1);
    if (e.key === 'ArrowRight') shift(1);
  });
})();
</script>

```

### templates/report.html

`$lang
<!-- templates/report.html -->
{{ template "layout.html" . }}

{{ define "report_content" }}
<div class="max-w-4xl mx-auto">
  <div class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
    <div class="bg-gradient-to-r from-slate-900 to-slate-800 px-6 py-8">
      <h2 class="text-2xl font-bold text-white">Report {{ if eq .type "lost" }}Lost{{ else }}Found{{ end }} Item</h2>
      <p class="mt-2 text-slate-300">Help reunite someone with their belongings</p>
    </div>

    <div class="p-6">
      {{ if .error }}
        <div class="mb-6 rounded-lg bg-rose-50 border border-rose-200 px-4 py-3 flex items-start">
          <svg class="h-5 w-5 text-rose-500 mr-3 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-rose-700">{{ .error }}</span>
        </div>
      {{ end }}

      <form method="POST" action="/report/new" enctype="multipart/form-data" class="space-y-6">
        <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
        <input type="hidden" name="type" value="{{ .type }}">

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Category *</label>
            <select name="category" required class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
              <option value="">Select Category</option>
              {{ range .categories }}<option value="{{ . }}">{{ . }}</option>{{ end }}
            </select>
          </div>
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Item Name *</label>
            <input type="text" name="title" required placeholder="e.g., HP Calculator" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Color</label>
            <select name="color" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
              <option value="">Select Color</option>
              {{ range .colors }}<option value="{{ . }}">{{ . }}</option>{{ end }}
            </select>
            <input type="text" name="color_other" placeholder="If other, specify color" class="w-full mt-2 px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
          </div>
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Brand</label>
            <input type="text" name="brand" placeholder="e.g., HP" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Location *</label>
            <select name="location" required class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
              <option value="">Select Location</option>
              {{ range .locations }}<option value="{{ . }}">{{ . }}</option>{{ end }}
            </select>
          </div>
          <div>
            <label class="block text-sm font-semibold text-slate-700 mb-2">Date Found/Lost</label>
            <input type="date" name="date" value="{{ now }}" max="{{ now }}" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
          </div>
        </div>

        <div>
          <label class="block text-sm font-semibold text-slate-700 mb-2">Description</label>
          <textarea name="description" rows="4" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent" placeholder="Describe your item in detail..."></textarea>
        </div>

        <div>
          <label class="block text-sm font-semibold text-slate-700 mb-2">Photos {{ if eq .type "found" }}*(required){{ end }}</label>
          <div class="upload-widget">
            <label for="images" class="upload-dropzone flex flex-col items-center justify-center">
              <svg class="h-10 w-10 text-slate-400 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p class="text-sm font-semibold text-slate-700">Drag & drop images here</p>
              <p class="text-xs text-slate-500">or</p>
              <span class="mt-2 inline-flex items-center justify-center px-4 py-2 rounded-lg bg-slate-900 text-white text-sm font-semibold shadow hover:bg-slate-800 transition">Choose files</span>
              <p class="text-xs text-slate-500 mt-2">Max 5MB per file. JPG or PNG. Required for found items.</p>
            </label>
            <input id="images" type="file" name="images" accept="image/*" multiple {{ if eq .type "found" }}required{{ end }} class="file-input-hidden">
            <div class="upload-preview mt-4"></div>
          </div>
        </div>

        <div class="pt-6 flex flex-col sm:flex-row gap-3">
          <button type="submit" class="flex-1 px-6 py-3 bg-gradient-to-r from-slate-900 to-slate-800 text-white rounded-lg font-semibold shadow-lg hover:shadow-xl transition-all duration-300 transform hover:-translate-y-0.5">Submit Report</button>
          <a href="/report" class="flex-1 px-6 py-3 bg-slate-100 text-slate-700 rounded-lg font-semibold hover:bg-slate-200 transition text-center">Cancel</a>
        </div>
      </form>
    </div>
  </div>
</div>
{{ end }}

```

### templates/dashboard.html

`$lang
<!-- templates/dashboard.html -->
{{ template "layout.html" . }}

{{ define "dashboard_content" }}
<div class="space-y-6">
  <div>
    <p class="text-sm text-slate-500">Welcome back</p>
    <h2 class="text-2xl font-bold text-slate-900">Hello, {{ .user.Name }}</h2>
  </div>

  <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
      <p class="text-xs uppercase tracking-wide text-slate-500">Lost</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_lost }}</p>
    </div>
    <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
      <p class="text-xs uppercase tracking-wide text-slate-500">Found</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_found }}</p>
    </div>
    <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
      <p class="text-xs uppercase tracking-wide text-slate-500">Claims</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_claims }}</p>
    </div>
    <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
      <p class="text-xs uppercase tracking-wide text-slate-500">Pending Claims</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.pending_claims }}</p>
    </div>
  </div>

  <div class="flex flex-wrap gap-3">
    <a href="/report/new?type=lost" class="bg-sky-600 text-white px-4 py-2 rounded-lg font-semibold shadow hover:-translate-y-[1px] transition">Report Lost</a>
    <a href="/report/new?type=found" class="bg-emerald-600 text-white px-4 py-2 rounded-lg font-semibold shadow hover:-translate-y-[1px] transition">Report Found</a>
    <a href="/report" class="bg-white text-slate-900 border border-slate-200 px-4 py-2 rounded-lg font-semibold shadow-sm hover:-translate-y-[1px] transition">Report View</a>
    <a href="/notifications" class="bg-white text-slate-900 border border-slate-200 px-4 py-2 rounded-lg font-semibold shadow-sm hover:-translate-y-[1px] transition">
      Notifications{{ if gt .unread_count 0 }} <span class="ml-2 inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full bg-rose-500 text-white text-[11px]">{{ .unread_count }}</span>{{ end }}
    </a>
  </div>

  <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
    <div class="px-4 py-3 border-b border-slate-200 flex items-center justify-between">
      <h3 class="text-lg font-semibold text-slate-900">My Reported Items</h3>
    </div>
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="bg-slate-50 text-slate-600">
          <tr>
            <th class="px-4 py-3 text-left font-semibold">Title</th>
            <th class="px-4 py-3 text-left font-semibold">Type</th>
            <th class="px-4 py-3 text-left font-semibold">Location</th>
            <th class="px-4 py-3 text-left font-semibold">Approval</th>
            <th class="px-4 py-3 text-left font-semibold">Claim Status</th>
            <th class="px-4 py-3 text-left font-semibold">Admin Remark</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          {{ range .my_items }}
          <tr class="hover:bg-slate-50">
            <td class="px-4 py-3">{{ .Title }}</td>
            <td class="px-4 py-3"><span class="px-2 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800">{{ .Type }}</span></td>
            <td class="px-4 py-3">{{ .Location }}</td>
            <td class="px-4 py-3"><span class="px-2 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800">{{ .ApprovalStatus }}</span></td>
            <td class="px-4 py-3"><span class="px-2 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800">{{ .Status }}</span></td>
            <td class="px-4 py-3 text-slate-600">{{ if .AdminRemarks }}{{ .AdminRemarks }}{{ else }}-{{ end }}</td>
          </tr>
          {{ else }}
          <tr>
            <td class="px-4 py-6 text-center text-slate-500" colspan="6">No items reported yet.</td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</div>
{{ end }}

```

### templates/login.html

`$lang
<!-- templates/login.html -->
{{ template "layout.html" . }}

{{ define "login_content" }}
<div class="auth-form">
    <h2>Login</h2>
    
    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}
    
    <form method="POST" action="/login">
        <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
        <div class="form-group">
            <label>Student ID</label>
            <input type="text" name="student_id" required placeholder="ugr/38923/18" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>
        
        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>
        
        <button type="submit" class="btn btn-primary">Login</button>
    </form>
    
    <p class="auth-link">Don't have an account? <a href="/register">Register</a></p>
    
    <div class="demo-credentials">
        <p><strong>One Login For All Users:</strong></p>
        <p>Enter your ID and password. Admin users are redirected to admin dashboard automatically.</p>
    </div>
</div>
{{ end }}


```

### templates/register.html

`$lang
<!-- templates/register.html -->
{{ template "layout.html" . }}

{{ define "register_content" }}
<div class="auth-form">
    <h2>Register</h2>

    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}

    <form method="POST" action="/register">
        <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
        <div class="form-group">
            <label>Full Name</label>
            <input type="text" name="name" required placeholder="John Doe" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <div class="form-group">
            <label>Student ID</label>
            <input type="text" name="student_id" required placeholder="ugr/38923/18" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <div class="form-group">
            <label>Phone Number</label>
            <input type="text" name="phone" required placeholder="09xxxxxxxx" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required minlength="6" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <button type="submit" class="btn btn-primary">Register</button>
    </form>

    <p class="auth-link">Already have an account? <a href="/login">Login</a></p>
</div>
{{ end }}

```

### templates/notifications.html

`$lang
{{ template "layout.html" . }}

{{ define "notifications_content" }}
<div class="space-y-4">
  <div class="flex items-center justify-between flex-wrap gap-3">
    <div>
      <h2 class="text-2xl font-bold text-slate-900">My Notifications</h2>
      <p class="text-slate-600 text-sm">Messages are shown in full blocks; copy contact numbers with one click.</p>
    </div>
    <form method="POST" action="/notifications/read" class="shrink-0">
      <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
      <button type="submit" class="px-4 py-2 rounded-lg bg-slate-900 text-white text-sm font-semibold hover:bg-slate-800 transition">Mark All Read</button>
    </form>
  </div>

  {{ if .notifications }}
    <div class="space-y-3">
      {{ range .notifications }}
      <div class="bg-white border border-slate-200 rounded-xl shadow-sm p-4 {{ if not .IsRead }}ring-2 ring-amber-200{{ end }}" data-notif-card>
        <div class="flex flex-wrap items-start gap-2">
          <div class="flex-1">
            <p class="text-xs text-slate-500">{{ .CreatedAt.Format "Jan 02, 2006 15:04" }}</p>
            <h3 class="text-base font-semibold text-slate-900">{{ .Title }}</h3>
          </div>
          <span class="px-3 py-1 rounded-full text-xs font-semibold {{ if .IsRead }}bg-emerald-100 text-emerald-800{{ else }}bg-amber-100 text-amber-800{{ end }}">
            {{ if .IsRead }}read{{ else }}new{{ end }}
          </span>
        </div>

        <div class="mt-3 grid gap-2" data-message-block>
          <div class="notif-message whitespace-pre-line text-sm text-slate-800 leading-relaxed">{{ .Message }}</div>
          <div class="flex flex-wrap gap-2 copy-targets"></div>
        </div>
      </div>
      {{ end }}
    </div>
  {{ else }}
    <div class="bg-white border border-slate-200 rounded-xl p-8 text-center text-slate-600">No notifications yet.</div>
  {{ end }}
</div>

<script>
(function() {
  const cards = document.querySelectorAll('[data-notif-card]');
  if (!cards) return;

  const phoneRegex = /Phone:\s*([+0-9\-\s]+)/gi;
  const nameRegex = /Name:\s*([^\n]+)/gi;

  cards.forEach(card => {
    const messageEl = card.querySelector('.notif-message');
    const targetsEl = card.querySelector('.copy-targets');
    if (!messageEl || !targetsEl) return;

    const text = messageEl.innerText;
    renderContactBlocks(card, text);

    const phones = [...text.matchAll(phoneRegex)];
    const names = [...text.matchAll(nameRegex)];
    const seen = new Set();

    phones.forEach((m, idx) => {
      const phone = (m[1] || '').trim();
      if (!phone || seen.has(phone)) return;
      seen.add(phone);
      const name = (names[idx] && names[idx][1]) ? names[idx][1].trim() : 'Contact';
      const btn = document.createElement('button');
      const firstName = name.split(' ')[0] || name;
      btn.type = 'button';
      btn.textContent = `Copy ${firstName}'s phone`;
      btn.className = 'px-3 py-1.5 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-slate-800 transition';
      btn.addEventListener('click', async () => {
        try {
          await navigator.clipboard.writeText(phone);
          btn.textContent = 'Copied!';
          setTimeout(() => { btn.textContent = `Copy ${firstName}'s phone`; }, 1500);
        } catch (err) {
          alert('Copy failed. Please copy manually: ' + phone);
        }
      });
      targetsEl.appendChild(btn);
    });
  });

  function renderContactBlocks(card, text) {
    const container = document.createElement('div');
    container.className = 'mt-3 flex flex-col md:flex-row gap-3';
    const sections = [];
    let current = null;
    const lines = text.split(/\r?\n/);
    lines.forEach(line => {
      const trimmed = line.trim();
      if (/^contact /i.test(trimmed)) {
        if (current) sections.push(current);
        current = { title: trimmed, body: [] };
        return;
      }
      if (/^your details/i.test(trimmed)) {
        if (current) sections.push(current);
        current = { title: trimmed, body: [] };
        return;
      }
      if (current) current.body.push(trimmed);
    });
    if (current) sections.push(current);

    if (sections.length === 0) return;

    sections.forEach(sec => {
      const cardEl = document.createElement('div');
      cardEl.className = 'flex-1 min-w-[240px] rounded-lg border border-slate-200 bg-slate-50 p-3';
      const titleEl = document.createElement('p');
      titleEl.className = 'text-xs font-semibold uppercase tracking-wide text-slate-500';
      titleEl.textContent = sec.title;
      cardEl.appendChild(titleEl);
      const bodyEl = document.createElement('div');
      bodyEl.className = 'mt-1 space-y-1 text-sm text-slate-800';
      sec.body.filter(Boolean).forEach(line => {
        const p = document.createElement('p');
        p.textContent = line;
        bodyEl.appendChild(p);
      });
      cardEl.appendChild(bodyEl);
      container.appendChild(cardEl);
    });

    const messageBlock = card.querySelector('[data-message-block]');
    if (messageBlock) {
      messageBlock.insertAdjacentElement('afterend', container);
    }
  }
})();
</script>
{{ end }}

```

### templates/error.html

`$lang
<!-- templates/error.html -->
{{ template "layout.html" . }}

{{ define "error_content" }}
<div class="error-page">
    <h2>{{ .title }}</h2>
    <div class="error">{{ .error }}</div>
    <a href="/" class="btn btn-secondary">Back Home</a>
</div>
{{ end }}

```

### templates/admin_login.html

`$lang
<!-- templates/admin_login.html -->
{{ template "layout.html" . }}

{{ define "admin_login_content" }}
<div class="auth-form">
    <h2>Admin Login</h2>
    

    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}

    <form method="POST" action="/admin/login">
        <input type="hidden" name="csrf_token" value="{{ .csrf_token }}">
        <div class="form-group">
            <label>Admin ID</label>
            <input type="text" name="student_id" required placeholder="admin" class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required class="w-full px-3 py-2 rounded-lg border border-slate-300 focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        </div>

        <button type="submit" class="btn btn-primary">Login as Admin</button>
    </form>
</div>
{{ end }}

```

### templates/admin/admin_dashboard.html

`$lang
<!-- templates/admin/dashboard.html -->
{{ template "layout.html" . }}

{{ define "admin_dashboard_content" }}
<div class="space-y-6">
  <div>
    <p class="text-sm text-slate-500">Admin overview</p>
    <h2 class="text-2xl font-bold text-slate-900">Admin Dashboard</h2>
  </div>

  <div class="grid sm:grid-cols-2 lg:grid-cols-5 gap-4">
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
      <p class="text-xs uppercase text-slate-500 font-semibold">Lost</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_lost }}</p>
    </div>
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
      <p class="text-xs uppercase text-slate-500 font-semibold">Found</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_found }}</p>
    </div>
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
      <p class="text-xs uppercase text-slate-500 font-semibold">Claims</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.total_claims }}</p>
    </div>
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
      <p class="text-xs uppercase text-slate-500 font-semibold">Pending Claims</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.pending_claims }}</p>
    </div>
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
      <p class="text-xs uppercase text-slate-500 font-semibold">Pending Posts</p>
      <p class="text-3xl font-semibold text-slate-900">{{ .stats.pending_items }}</p>
    </div>
  </div>

  <div class="flex flex-wrap gap-3">
    <a href="/admin/claims" class="bg-sky-600 text-white px-4 py-2 rounded-lg text-sm font-semibold shadow hover:-translate-y-[1px] transition">Manage Claims</a>
    <a href="/admin/items" class="bg-slate-900 text-white px-4 py-2 rounded-lg text-sm font-semibold shadow hover:-translate-y-[1px] transition">Manage Posts</a>
    <a href="/report/new?type=lost" class="bg-white border border-slate-300 text-slate-900 px-4 py-2 rounded-lg text-sm font-semibold hover:-translate-y-[1px] transition">Report Lost</a>
    <a href="/report/new?type=found" class="bg-white border border-slate-300 text-slate-900 px-4 py-2 rounded-lg text-sm font-semibold hover:-translate-y-[1px] transition">Report Found</a>
    <a href="/report" class="bg-white border border-slate-300 text-slate-900 px-4 py-2 rounded-lg text-sm font-semibold hover:-translate-y-[1px] transition">Report View</a>
  </div>

  <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
    <div class="px-4 py-3 border-b border-slate-200 flex items-center justify-between">
      <h3 class="text-lg font-semibold text-slate-900">Recent Claims</h3>
    </div>
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="bg-slate-50 text-slate-600">
          <tr>
            <th class="px-4 py-3 text-left font-semibold">ID</th>
            <th class="px-4 py-3 text-left font-semibold">Item</th>
            <th class="px-4 py-3 text-left font-semibold">Claimant</th>
            <th class="px-4 py-3 text-left font-semibold">Status</th>
            <th class="px-4 py-3 text-left font-semibold">Date</th>
            <th class="px-4 py-3 text-left font-semibold">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          {{ range .claims }}
          <tr class="hover:bg-slate-50">
            <td class="px-4 py-3">{{ .ID }}</td>
            <td class="px-4 py-3">
              <div class="font-semibold text-slate-900">{{ .Item.Title }}</div>
              <div class="text-xs text-slate-500">{{ .Item.Type }} • {{ .Item.Category }}</div>
            </td>
            <td class="px-4 py-3">{{ .User.Name }}</td>
            <td class="px-4 py-3"><span class="px-2 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800">{{ .Status }}</span></td>
            <td class="px-4 py-3">{{ .CreatedAt.Format "Jan 02, 2006" }}</td>
            <td class="px-4 py-3"><a href="/admin/claims" class="text-sky-700 font-semibold text-sm">View</a></td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</div>
{{ end }}

```

### templates/admin/admin_claims.html

`$lang
<!-- templates/admin/claims.html -->
{{ template "layout.html" . }}

{{ define "admin_claims_content" }}
<div class="space-y-6">
  <div class="flex items-center justify-between flex-wrap gap-2">
    <div>
      <h2 class="text-2xl font-bold text-slate-900">Manage Item Requests</h2>
      <p class="text-slate-600 text-sm">Review the original post and the requester side-by-side before approving.</p>
    </div>
  </div>

  {{ if .claims }}
    <div class="space-y-5">
      {{ range .claims }}
      <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
        <div class="border-b border-slate-100 px-6 py-4 flex flex-wrap items-center gap-3">
          <span class="px-3 py-1 rounded-full text-xs font-semibold {{ if eq .RequestType "found_match_request" }}bg-amber-100 text-amber-800{{ else }}bg-sky-100 text-sky-800{{ end }}">
            {{ if eq .RequestType "found_match_request" }}Found Match Request{{ else }}Claim Request{{ end }}
          </span>
          <span class="px-3 py-1 rounded-full text-xs font-semibold {{ if eq .Status "pending" }}bg-yellow-100 text-yellow-800{{ else if eq .Status "approved" }}bg-emerald-100 text-emerald-800{{ else }}bg-rose-100 text-rose-800{{ end }}">
            {{ .Status }}
          </span>
          <span class="text-xs text-slate-500">Submitted {{ .CreatedAt.Format "Jan 02, 2006" }}</span>
          <span class="ml-auto text-xs font-semibold text-slate-700 px-2 py-1 bg-slate-100 rounded">Post: {{ .Item.Type }}</span>
        </div>

        <div class="grid lg:grid-cols-2">
          <!-- Original post -->
          <div class="p-6 border-b lg:border-b-0 lg:border-r border-slate-100 bg-slate-50">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Original Post</p>
                <h3 class="text-lg font-bold text-slate-900 leading-tight">{{ .Item.Title }}</h3>
                <p class="text-sm text-slate-600">{{ .Item.Category }} • {{ .Item.Color }} • {{ .Item.Location }}</p>
              </div>
              <span class="px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-900 text-white">{{ .Item.ApprovalStatus }}</span>
            </div>

            <div class="mt-4">
              <div class="h-40 rounded-xl border border-slate-200 bg-white flex items-center justify-center overflow-hidden">
                {{ if .Item.Image }}
                  <img src="{{ .Item.Image }}" alt="{{ .Item.Title }}" class="max-h-full max-w-full object-contain p-2">
                {{ else if gt (len .Item.Images) 0 }}
                  <img src="{{ (index .Item.Images 0).Path }}" alt="{{ .Item.Title }}" class="max-h-full max-w-full object-contain p-2">
                {{ else }}
                  <span class="text-sm text-slate-400">No image</span>
                {{ end }}
              </div>
            </div>

            <dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-sm text-slate-700">
              <div><dt class="font-semibold text-slate-600">Brand</dt><dd>{{ if .Item.Brand }}{{ .Item.Brand }}{{ else }}—{{ end }}</dd></div>
              <div><dt class="font-semibold text-slate-600">Date</dt><dd>{{ if .Item.Date }}{{ .Item.Date }}{{ else }}—{{ end }}</dd></div>
              <div class="col-span-2"><dt class="font-semibold text-slate-600">Status</dt><dd class="capitalize">{{ .Item.Status }}</dd></div>
              <div class="col-span-2 flex items-center gap-3 flex-wrap">
                <div>
                  <dt class="font-semibold text-slate-600">Posted By</dt>
                  <dd>{{ .Item.User.Name }} ({{ .Item.User.StudentID }}) • {{ .Item.User.Phone }}</dd>
                </div>
                <button type="button" class="px-3 py-1.5 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-slate-800 transition copy-phone" data-phone="{{ .Item.User.Phone }}">Copy poster phone</button>
              </div>
              <div class="col-span-2"><dt class="font-semibold text-slate-600">Description</dt><dd class="text-slate-700 mt-1 leading-relaxed">{{ if .Item.Description }}{{ .Item.Description }}{{ else }}No description provided.{{ end }}</dd></div>
            </dl>
          </div>

          <!-- Requester side -->
          <div class="p-6 space-y-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Requester</p>
                <h3 class="text-lg font-bold text-slate-900 leading-tight">{{ .User.Name }}</h3>
                <p class="text-sm text-slate-600">{{ .User.StudentID }} • {{ .User.Phone }}</p>
              </div>
              <span class="text-xs text-slate-500">Updated {{ .UpdatedAt.Format "Jan 02, 2006" }}</span>
            </div>

            <div class="rounded-xl border border-slate-200 bg-slate-50 p-4">
              <p class="text-xs font-semibold text-slate-500 uppercase tracking-wide">Message</p>
              <p class="text-sm text-slate-700 mt-1 leading-relaxed">{{ if .Description }}{{ .Description }}{{ else }}No message provided.{{ end }}</p>
            </div>

            <div class="flex items-center gap-3 flex-wrap">
              <div class="text-sm text-slate-700">
                <p class="font-semibold text-slate-600">Requester Contact</p>
                <p>{{ .User.Name }} ({{ .User.StudentID }}) • {{ .User.Phone }}</p>
              </div>
              <button type="button" class="px-3 py-1.5 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-slate-800 transition copy-phone" data-phone="{{ .User.Phone }}">Copy requester phone</button>
            </div>

            {{ if eq .Status "pending" }}
              <form method="POST" action="/admin/claims/update" class="space-y-3">
                <input type="hidden" name="csrf_token" value="{{ $.csrf_token }}">
                <input type="hidden" name="claim_id" value="{{ .ID }}">
                <label class="block text-sm font-semibold text-slate-700">Remarks to both parties (optional)</label>
                <input type="text" name="remarks" placeholder="e.g., Please meet at the main gate at 3 PM" class="w-full px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
                <div class="flex flex-wrap gap-2">
                  <button type="submit" name="status" value="approved" class="px-4 py-2 rounded-lg bg-emerald-600 text-white text-sm font-semibold shadow hover:bg-emerald-700 transition">Approve & notify both</button>
                  <button type="submit" name="status" value="rejected" class="px-4 py-2 rounded-lg bg-rose-600 text-white text-sm font-semibold shadow hover:bg-rose-700 transition">Reject & notify requester</button>
                </div>
              </form>
            {{ else }}
              <div class="rounded-xl border border-slate-200 bg-slate-50 p-4 space-y-2">
                <p class="text-sm font-semibold text-slate-700">Admin decision: <span class="capitalize">{{ .Status }}</span></p>
                <p class="text-sm text-slate-600">Remarks: {{ if .AdminRemarks }}{{ .AdminRemarks }}{{ else }}No remarks provided.{{ end }}</p>
              </div>
            {{ end }}
          </div>
        </div>
      </div>
      {{ end }}
    </div>
  {{ else }}
    <div class="bg-white border border-slate-200 rounded-xl p-8 text-center text-slate-600">No claims yet.</div>
  {{ end }}
</div>

<script>
(function() {
  document.querySelectorAll('.copy-phone').forEach(btn => {
    const phone = (btn.getAttribute('data-phone') || '').trim();
    if (!phone) {
      btn.disabled = true;
      btn.classList.add('opacity-60');
      return;
    }
    btn.addEventListener('click', async () => {
      try {
        await navigator.clipboard.writeText(phone);
        const original = btn.textContent;
        btn.textContent = 'Copied!';
        setTimeout(() => { btn.textContent = original; }, 1500);
      } catch (err) {
        alert('Copy failed. Please copy manually: ' + phone);
      }
    });
  });
})();
</script>
{{ end }}

```

### templates/admin/admin_items.html

`$lang
<!-- templates/admin/admin_items.html -->
{{ template "layout.html" . }}

{{ define "admin_items_content" }}
<div class="space-y-6">
  <div class="flex items-center justify-between flex-wrap gap-2">
    <h2 class="text-2xl font-bold text-slate-900">Manage Item Posts</h2>
  </div>

  <form method="GET" action="/admin/items" class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm">
    <div class="flex flex-wrap gap-3">
      <input type="text" name="q" value="{{ .selected_q }}" placeholder="Keyword" class="w-56 px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
      <select name="type" class="w-40 px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        <option value="">All types</option>
        <option value="lost" {{ if eq .selected_type "lost" }}selected{{ end }}>Lost</option>
        <option value="found" {{ if eq .selected_type "found" }}selected{{ end }}>Found</option>
      </select>
      <select name="status" class="w-40 px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        <option value="">All approvals</option>
        <option value="pending" {{ if eq .selected_status "pending" }}selected{{ end }}>Pending</option>
        <option value="approved" {{ if eq .selected_status "approved" }}selected{{ end }}>Approved</option>
        <option value="rejected" {{ if eq .selected_status "rejected" }}selected{{ end }}>Rejected</option>
      </select>
      <select name="category" class="w-44 px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        <option value="">All categories</option>
        {{ range .categories }}
        <option value="{{ . }}" {{ if eq $.selected_category . }}selected{{ end }}>{{ . }}</option>
        {{ end }}
      </select>
      <select name="location" class="w-44 px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
        <option value="">All locations</option>
        {{ range .locations }}
        <option value="{{ . }}" {{ if eq $.selected_location . }}selected{{ end }}>{{ . }}</option>
        {{ end }}
      </select>
      <input type="date" name="date_from" value="{{ .selected_date_from }}" max="{{ now }}" class="px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
      <input type="date" name="date_to" value="{{ .selected_date_to }}" max="{{ now }}" class="px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
    </div>
    <div class="mt-3 flex flex-wrap gap-2">
      <button type="submit" class="bg-slate-900 text-white px-4 py-2 rounded-lg text-sm font-semibold shadow hover:-translate-y-[1px] transition">Filter</button>
      <a href="/admin/items" class="bg-white border border-slate-300 text-slate-800 px-4 py-2 rounded-lg text-sm font-semibold hover:-translate-y-[1px] transition">Clear</a>
    </div>
  </form>

  <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="bg-slate-50 text-slate-600">
          <tr>
            <th class="px-4 py-3 text-left font-semibold">ID</th>
            <th class="px-4 py-3 text-left font-semibold">Item</th>
            <th class="px-4 py-3 text-left font-semibold">Details</th>
            <th class="px-4 py-3 text-left font-semibold">Photos</th>
            <th class="px-4 py-3 text-left font-semibold">Reporter</th>
            <th class="px-4 py-3 text-left font-semibold">Approval</th>
            <th class="px-4 py-3 text-left font-semibold">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          {{ range .items }}
          <tr class="hover:bg-slate-50">
            <td class="px-4 py-3">{{ .ID }}</td>
            <td class="px-4 py-3">
              <div class="font-semibold text-slate-900">{{ .Title }}</div>
              <div class="text-xs text-slate-500">Type: {{ .Type }}</div>
            </td>
            <td class="px-4 py-3 text-slate-700 text-xs leading-5">
              Category: {{ .Category }}<br>
              Color: {{ .Color }}<br>
              Brand: {{ if .Brand }}{{ .Brand }}{{ else }}-{{ end }}<br>
              Location: {{ .Location }}<br>
              Date: {{ .Date }}<br>
              Desc: {{ if .Description }}{{ .Description }}{{ else }}-{{ end }}
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-2 flex-wrap">
              {{ if .Images }}
                {{ range .Images }}
                <img src="{{ .Path }}" class="w-14 h-14 object-cover rounded border border-slate-200">
                {{ end }}
              {{ else if .Image }}
                <img src="{{ .Image }}" class="w-14 h-14 object-cover rounded border border-slate-200">
              {{ else }}<span class="text-slate-500 text-xs">No image</span>{{ end }}
              </div>
            </td>
            <td class="px-4 py-3 text-xs text-slate-700">
              {{ .User.Name }}<br><span class="text-slate-500">{{ .User.StudentID }} | {{ .User.Phone }}</span>
            </td>
            <td class="px-4 py-3"><span class="px-2 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-800">{{ .ApprovalStatus }}</span></td>
            <td class="px-4 py-3">
              <form method="POST" action="/admin/items/update" class="flex flex-col gap-2">
                <input type="hidden" name="csrf_token" value="{{ $.csrf_token }}">
                <input type="hidden" name="item_id" value="{{ .ID }}">
                <input type="text" name="remarks" placeholder="Remarks" class="px-3 py-2 rounded-lg border border-slate-300 text-sm focus:ring-2 focus:ring-sky-500 focus:border-transparent">
                <div class="flex gap-2">
                  <button type="submit" name="approval_status" value="approved" class="bg-emerald-600 text-white px-3 py-1 rounded-lg text-xs font-semibold">Approve</button>
                  <button type="submit" name="approval_status" value="rejected" class="bg-rose-600 text-white px-3 py-1 rounded-lg text-xs font-semibold">Reject</button>
                </div>
              </form>
              <form method="POST" action="/admin/items/delete" class="mt-2">
                <input type="hidden" name="csrf_token" value="{{ $.csrf_token }}">
                <input type="hidden" name="item_id" value="{{ .ID }}">
                <button type="submit" class="bg-white border border-rose-300 text-rose-700 px-3 py-1 rounded-lg text-xs font-semibold">Remove</button>
              </form>
            </td>
          </tr>
          {{ else }}
          <tr>
            <td class="px-4 py-6 text-center text-slate-500" colspan="7">No posts found.</td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</div>
{{ end }}

```
