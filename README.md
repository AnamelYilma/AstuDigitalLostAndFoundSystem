## This Is ASTU Steam Project BY Anamel Yilma .Simple Startage But Effective way

# 🧠 1. Big Architecture View (Simple)

You are building a classic **MVC-style backend web app**:

```
Browser (User)
     ↓
Gin Router
     ↓
Handlers (Controllers)
     ↓
Services (Business Logic)
     ↓
Repository (Database Access)
     ↓
Database
```

HTML is rendered from Gin.

---

# 📁 2. Project Folder Structure (Production Style)

This is what I expect from you:
```
├── internal
│   ├── handler
│   │   ├── admin_handler.go
│   │   ├── auth_handler.go
│   │   └── item_handler.go
│   ├── middleware
│   │   └── auth_middleware.go
│   ├── model
│   │   ├── item.go
│   │   └── user.go
│   ├── repository
│   │   ├── item_repository.go
│   │   └── user_repository.go
│   └── service
│       ├── auth_service.go
│       ├── item_options.go
│       └── item_service.go
├── pkg
│   ├── database
│   │   └── db.go
│   └── utils
│       └── hash.go
├── static
│   ├── css
│   │   └── style.css
│   └── js
├── templates
│   ├── admin
│   │   ├── claims.html
│   │   └── dashboard.html
│   ├── dashboard.html
│   ├── error.html
│   ├── index.html
│   ├── item.html
│   ├── items.html
│   ├── layout.html
│   ├── login.html
│   ├── register.html
│   ├── report.html
│   └── search.html
├── go.mod
├── go.sum
├── main.go
```

If you don’t follow structure like this, you are building toy project.

---

# 🔥 3. What Each Folder Means (Understand This Deeply)

## `main.go`

Entry point.

* Initialize database
* Setup Gin
* Register routes
* Run server

This is where everything starts.

---

## `/internal/model`

Database models (structs).

Example:

```go
type Hotel struct {
    ID       uint
    Name     string
    Location string
    Approved bool
}
```

Only structure. No logic.

---

## `/internal/repository`

Only talks to database.

Example:

* CreateHotel()
* GetHotels()
* SearchHotel()
* ApproveHotel()

Repository = database layer only.

No business logic.

---

## `/internal/service`

Business logic lives here.

Example:

* If hotel not approved → don’t show to public
* Only admin can approve
* Search + filter rules

This is where brain logic lives.

---

## `/internal/handler`

Gin handlers.

Example:

```go
func (h *HotelHandler) List(c *gin.Context)
```

This:

* Read query parameters
* Call service
* Render HTML template

Handlers do NOT contain logic.
They only connect HTTP to service.

---

## `/templates`

Gin renders HTML from here.

Yes, Gin can render HTML.

Example in main.go:

```go
router.LoadHTMLGlob("templates/*")
router.GET("/", handler.Home)
```

Inside handler:

```go
c.HTML(200, "index.html", gin.H{
    "hotels": hotels,
})
```

So yes — no React needed.

---

## `/middleware`

Authentication.

Example:

* Check if user logged in
* Check role (admin or normal user)

Without middleware → security is trash.

---

# 🔐 4. Security Structure You Must Use

Minimum:

### 1. Password hashing

Use bcrypt.

### 2. Session-based auth

Use cookies.

### 3. Role-based access

Admin routes:

```
/admin/*
```

Protected by middleware.

### 4. Input validation

Never trust user input.

---

# ⚙️ 5. How Features Work Logically

## 🔍 Searching

Frontend:

```
/hotels?name=addis&location=debre
```

Handler:

* Read query
* Pass to service
* Service builds filter
* Repository builds SQL WHERE

---

## 🧮 Filtering

Example:

* Price range
* Location
* Rating

Same logic as search.

---

## 📊 Reporting

Admin dashboard:

* Total hotels
* Approved hotels
* Pending hotels
* Users count

Service calls repository COUNT queries.

---

## ✅ Approval Flow

1. User submits hotel
2. Approved = false
3. Admin sees pending list
4. Admin clicks approve
5. Update Approved = true

Only approved hotels show to public.

---

# 🏗 6. How Code Flow Works (Real Flow Example)

User opens homepage:

```
Browser
  ↓
GET /
  ↓
Handler.Home()
  ↓
Service.GetApprovedHotels()
  ↓
Repository.FindApproved()
  ↓
Database
  ↓
Return data
  ↓
Render HTML
```



Be precise.
