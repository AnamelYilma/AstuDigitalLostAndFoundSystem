# 📚 COMPLETE GO CODE DOCUMENTATION WITH DETAILED COMMENTS

## 📁 Project Structure

```
ASTU_Lost_and_found_system/
├── main.go                    → Application entry point
├── internal/
│   ├── handler/              → HTTP request handlers (controllers)
│   │   ├── auth_handler.go   → Login, register, logout
│   │   ├── item_handler.go   → Items, claims, search
│   │   ├── admin_handler.go  → Admin operations
│   │   └── render.go         → HTML rendering helper
│   ├── middleware/           → HTTP middleware (interceptors)
│   │   ├── auth_middleware.go → Authentication & authorization
│   │   └── csrf_middleware.go → CSRF protection
│   ├── model/                → Database models (tables)
│   │   ├── user.go           → User table structure
│   │   ├── item.go           → Item, Claim, ItemImage tables
│   │   └── notification.go   → Notification table
│   ├── repository/           → Database operations (queries)
│   │   ├── user_repository.go → User database queries
│   │   └── item_repository.go → Item database queries
│   └── service/              → Business logic
│       ├── auth_service.go   → Authentication logic
│       ├── item_service.go   → Item management logic
│       └── item_options.go   → Dropdown options & validation
└── pkg/
    ├── database/
    │   └── db.go             → Database connection
    └── utils/
        └── hash.go           → Password hashing
```

---

# 1️⃣ MAIN.GO - Application Entry Point

```go
// Package declaration - this file belongs to the main package
// The main package is special in Go - it creates an executable program
package main

// Import statements - bring in external libraries and packages
import (
	"bufio"                          // For reading files line by line
	"fmt"                            // For string formatting (Sprintf, Printf)
	"html/template"                  // For HTML template rendering
	"log"                            // For logging messages
	"lostfound/internal/handler"     // Our HTTP handlers
	"lostfound/internal/middleware"  // Our middleware functions
	"lostfound/internal/model"       // Our database models
	"lostfound/pkg/database"         // Database connection
	"lostfound/pkg/utils"            // Utility functions
	"net/http"                       // HTTP constants and types
	"os"                             // Operating system functions
	"strings"                        // String manipulation
	"time"                           // Time and date functions
	"github.com/gin-gonic/gin"       // Gin web framework
)

// ========== MAIN FUNCTION - PROGRAM STARTS HERE ==========
func main() {
	// ========== STEP 1: LOAD CONFIGURATION ==========
	
	// Load environment variables from .env file
	// This reads DB_HOST, DB_PASSWORD, PORT, etc.
	loadDotEnv(".env")
	
	// Get PORT from environment variable
	// On Render: They set this automatically (e.g., PORT=10000)
	// Locally: If not set, we default to 8080
	port := os.Getenv("PORT")
	if port == "" { 
		port = "8080"  // Default port for local development
	}
	address := ":" + port  // Creates ":8080" or ":10000" format
	
	// Check if running in production mode
	// strings.EqualFold = case-insensitive comparison
	if strings.EqualFold(os.Getenv("GO_ENV"), "production") {
		// Set Gin to release mode (no debug logs, better performance)
		gin.SetMode(gin.ReleaseMode)
	}

	// ========== STEP 2: DATABASE SETUP ==========
	
	// Connect to PostgreSQL database
	// Uses environment variables: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT
	database.InitDB()
	
	// AutoMigrate: Automatically create/update database tables
	// &model.User{} = pointer to User struct (tells GORM to create users table)
	// If table doesn't exist, it creates it
	// If columns are missing, it adds them
	database.DB.AutoMigrate(
		&model.User{},         // Creates 'users' table
		&model.Item{},         // Creates 'items' table
		&model.ItemImage{},    // Creates 'item_images' table
		&model.Claim{},        // Creates 'claims' table
		&model.Notification{}, // Creates 'notifications' table
	)
	
	// Create default admin account if none exists
	// Default credentials: username=admin, password=admin123
	createDefaultAdmin()
	
	// Fix old/legacy data from previous versions
	// Ensures all users have student_id and phone
	normalizeLegacyData()
	
	// Add database constraints (unique indexes, NOT NULL)
	enforceUserConstraints()

	// ========== STEP 3: WEB SERVER SETUP ==========
	
	// Create Gin router with default middleware
	// Default middleware includes: Logger (logs requests) and Recovery (handles panics)
	r := gin.Default()
	
	// Create custom template function map
	// This allows us to use custom functions in HTML templates
	funcMap := template.FuncMap{
		// Define "now" function that returns current date
		// Usage in HTML: {{now}} will display "2024-01-15"
		"now": func() string {
			// "2006-01-02" is Go's reference time format (YYYY-MM-DD)
			return time.Now().Format("2006-01-02")
		},
	}

	// Load HTML templates
	tmpl := template.New("").Funcs(funcMap)  // Create new template with custom functions
	
	// template.Must = panic if error occurs (ensures templates load correctly)
	// ParseGlob = load all files matching pattern
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))        // Load main templates
	tmpl = template.Must(tmpl.ParseGlob("templates/admin/*.html"))  // Load admin templates
	
	// Register templates with Gin router
	r.SetHTMLTemplate(tmpl)

	// ========== STEP 4: MIDDLEWARE SETUP ==========
	
	// Serve static files (CSS, JavaScript, images)
	// http.Dir("static") = serve files from "static" folder
	// URL /static/css/style.css → serves file static/css/style.css
	r.StaticFS("/static", http.Dir("static"))
	
	// Add SetUser middleware to ALL routes
	// This checks if user is logged in and adds user info to request context
	r.Use(middleware.SetUser())
	
	// Add CSRF middleware to ALL routes
	// This generates security tokens to prevent cross-site request forgery attacks
	r.Use(middleware.CSRFMiddleware())

	// ========== STEP 5: CREATE HANDLERS ==========
	
	// Handlers contain the logic for each route
	authHandler := handler.NewAuthHandler()    // Handles login, register, logout
	itemHandler := handler.NewItemHandler()    // Handles items, claims, search
	adminHandler := handler.NewAdminHandler()  // Handles admin operations

	// ========== STEP 6: PUBLIC ROUTES (No login required) ==========
	
	// Home page route "/"
	// r.GET = handle GET requests
	// func(c *gin.Context) = anonymous function (inline handler)
	r.GET("/", func(c *gin.Context) {
		// Get user from context (set by SetUser middleware)
		// _ = ignore error (we don't care if user doesn't exist)
		user, _ := c.Get("user")
		
		// Check if user is logged in
		// u, ok := user.(model.User) = type assertion
		// ok = true if user is actually a User type
		if u, ok := user.(model.User); ok {
			// User is logged in
			
			// If user is admin, redirect to admin dashboard
			if u.Role == "admin" {
				c.Redirect(303, "/admin/dashboard")  // 303 = See Other (POST redirect)
				return  // Stop execution
			}
			
			// If regular user, redirect to user dashboard
			c.Redirect(303, "/dashboard")
			return
		}
		
		// User is NOT logged in, show home page
		
		// Get CSRF token from context
		csrfToken, _ := c.Get("csrf_token")
		
		// Declare variable to hold unread notification count
		var unread int64
		
		// Render HTML template
		// 200 = HTTP 200 OK status
		// "index.html" = template file name
		// gin.H = shortcut for map[string]interface{}
		c.HTML(200, "index.html", gin.H{
			"title":            "ASTU Lost & Found",  // Page title
			"user":             user,                  // User object (nil if not logged in)
			"unread_count":     unread,                // Notification count
			"csrf_token":       csrfToken,             // CSRF token for forms
			"content_template": "index_content",       // Sub-template to include
		})
	})

	// Authentication routes
	r.GET("/login", authHandler.ShowLogin)       // GET /login → show login page
	r.POST("/login", authHandler.Login)          // POST /login → process login form
	r.GET("/register", authHandler.ShowRegister) // GET /register → show registration page
	r.POST("/register", authHandler.Register)    // POST /register → process registration
	r.GET("/logout", authHandler.Logout)         // GET /logout → logout user

	// Public item viewing routes
	r.GET("/report", itemHandler.Search)  // Search/browse lost & found items
	
	// Redirect old URLs to /report
	r.GET("/items", func(c *gin.Context) { c.Redirect(303, "/report") })
	r.GET("/search", func(c *gin.Context) { c.Redirect(303, "/report") })
	
	// View single item details
	// :id = URL parameter (e.g., /item/123 → id=123)
	r.GET("/item/:id", itemHandler.ShowItem)

	// ========== STEP 7: PROTECTED ROUTES (Login required) ==========
	
	// Create route group with base path "/"
	protected := r.Group("/")
	
	// Add AuthRequired middleware to this group
	// All routes in this group will check if user is logged in
	protected.Use(middleware.AuthRequired())
	
	// { } = group block (all routes inside require authentication)
	{
		protected.GET("/dashboard", itemHandler.Dashboard)                      // User dashboard
		protected.GET("/report/new", itemHandler.ShowReportForm)                // Show report form
		protected.POST("/report/new", itemHandler.ReportItem)                   // Submit report
		protected.POST("/claim", itemHandler.ClaimItem)                         // Claim an item
		protected.GET("/notifications", itemHandler.ShowNotifications)          // View notifications
		protected.POST("/notifications/read", itemHandler.MarkNotificationsRead) // Mark as read
	}

	// ========== STEP 8: ADMIN ROUTES (Admin role required) ==========
	
	// Create admin route group with base path "/admin"
	admin := r.Group("/admin")
	
	// Add TWO middleware functions
	admin.Use(middleware.AuthRequired())  // 1. Check if logged in
	admin.Use(middleware.AdminRequired()) // 2. Check if user is admin
	
	{
		admin.GET("/dashboard", adminHandler.Dashboard)         // Admin dashboard
		admin.GET("/claims", adminHandler.ShowClaims)           // View all claims
		admin.POST("/claims/update", adminHandler.UpdateClaim)  // Approve/reject claim
		admin.GET("/items", adminHandler.ShowItems)             // View all items
		admin.POST("/items/update", adminHandler.UpdateItem)    // Update item status
		admin.POST("/items/delete", adminHandler.DeleteItem)    // Delete item
	}

	// ========== STEP 9: START THE SERVER ==========
	
	// Log message to console
	// %s = string placeholder (replaced with address value)
	log.Printf("Server starting on %s", address)
	
	// Start the HTTP server
	// r.Run(address) = start server on specified address
	// This function blocks here (waits for requests)
	if err := r.Run(address); err != nil {
		// If server fails to start, log error and exit
		log.Fatal(err)
	}
}

// ========== HELPER FUNCTION: LOAD .ENV FILE ==========
func loadDotEnv(path string) {
	// Open the .env file
	// os.Open = open file for reading
	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, just return (not fatal)
		return
	}
	// defer = execute this when function ends
	// Ensures file is closed even if error occurs
	defer file.Close()

	// Create scanner to read file line by line
	scanner := bufio.NewScanner(file)
	
	// Loop through each line
	// scanner.Scan() = read next line, returns false when done
	for scanner.Scan() {
		// Get current line text
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue  // Skip to next iteration
		}

		// Split line by "=" into key and value
		// SplitN(line, "=", 2) = split into max 2 parts
		// Example: "DB_HOST=localhost" → ["DB_HOST", "localhost"]
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue  // Invalid line, skip it
		}

		// Extract key and value
		key := strings.TrimSpace(parts[0])
		// Trim quotes from value
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		
		if key == "" {
			continue  // Empty key, skip
		}

		// Check if environment variable already exists
		// os.LookupEnv returns (value, exists)
		// _ = ignore value, we only care if it exists
		if _, exists := os.LookupEnv(key); !exists {
			// Variable doesn't exist, set it
			// _ = ignore error
			_ = os.Setenv(key, value)
		}
	}
}

// ========== HELPER FUNCTION: CREATE DEFAULT ADMIN ==========
func createDefaultAdmin() {
	// Declare variable to hold count
	// int64 = 64-bit integer
	var count int64
	
	// Count how many admin users exist
	// database.DB = global database connection
	// Model(&model.User{}) = specify we're querying users table
	// Where("role = ?", "admin") = filter where role equals "admin"
	// Count(&count) = count matching rows and store in count variable
	// ? = placeholder (prevents SQL injection)
	database.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	// If no admin exists, create one
	if count == 0 {
		// Hash the password "admin123"
		// utils.HashPassword = bcrypt hashing function
		// _ = ignore error (we trust it won't fail)
		hashedPassword, _ := utils.HashPassword("admin123")
		
		// Create admin user struct
		// &model.User = pointer to User struct
		admin := &model.User{
			Name:      "Admin",
			StudentID: "admin",
			Phone:     "0000000000",
			Email:     "admin@astu.edu",
			Password:  hashedPassword,  // Store hashed password, NOT plain text
			Role:      "admin",
		}
		
		// Save admin to database
		// Create = INSERT INTO users ...
		database.DB.Create(admin)
		
		// Log success message
		log.Println("Default admin created: admin / admin123")
		return  // Exit function
	}

	// Admin already exists, check if it needs updating
	
	// Declare variable to hold existing admin
	var existingAdmin model.User
	
	// Get first admin from database
	// First(&existingAdmin) = SELECT * FROM users WHERE role='admin' LIMIT 1
	// .Error = get any error that occurred
	if err := database.DB.Where("role = ?", "admin").First(&existingAdmin).Error; err == nil {
		// err == nil means NO error (admin was found)
		
		// Flag to track if we need to save changes
		needsUpdate := false
		
		// Check if StudentID is missing
		if existingAdmin.StudentID == "" {
			existingAdmin.StudentID = "admin"  // Set default
			needsUpdate = true                 // Mark for saving
		}
		
		// Check if Phone is missing
		if existingAdmin.Phone == "" {
			existingAdmin.Phone = "0000000000"  // Set default
			needsUpdate = true                  // Mark for saving
		}
		
		// If anything changed, save to database
		if needsUpdate {
			// Save = UPDATE users SET ... WHERE id = existingAdmin.ID
			database.DB.Save(&existingAdmin)
		}
	}
}

// ========== HELPER FUNCTION: FIX OLD DATA ==========
func normalizeLegacyData() {
	// ========== PART 1: FIX USER DATA ==========
	
	// Declare slice (array) to hold all users
	// []model.User = slice of User structs
	var users []model.User
	
	// Get all users from database, sorted by ID
	// Order("id ASC") = sort by ID ascending (1, 2, 3...)
	// Find(&users) = SELECT * FROM users ORDER BY id ASC
	// .Error = get any error
	if err := database.DB.Order("id ASC").Find(&users).Error; err == nil {
		// err == nil = success
		
		// Create map to track which StudentIDs we've seen
		// map[string]bool = key is string, value is bool
		// This prevents duplicate StudentIDs
		seen := map[string]bool{}
		
		// Loop through each user
		// range users = iterate over users slice
		// _ = ignore index (we don't need it)
		// u = current user
		for _, u := range users {
			// Flag to track if this user needs saving
			needsUpdate := false
			
			// ========== FIX MISSING STUDENT ID ==========
			
			// Clean up StudentID: lowercase and remove spaces
			// strings.ToLower = convert to lowercase
			// strings.TrimSpace = remove leading/trailing spaces
			sid := strings.ToLower(strings.TrimSpace(u.StudentID))
			
			// If StudentID is empty, create one
			if sid == "" {
				// Try to use email prefix (part before @)
				// strings.Split(email, "@") = split by @
				// [0] = get first part
				// Example: "alice@astu.edu" → "alice"
				baseID := strings.Split(strings.TrimSpace(u.Email), "@")[0]
				
				// If email is also empty, use user ID
				if baseID == "" {
					// fmt.Sprintf = format string
					// %d = integer placeholder
					baseID = fmt.Sprintf("user_%d", u.ID)
					// Example: "user_5"
				}
				
				sid = strings.ToLower(baseID)
				needsUpdate = true  // Mark for saving
			}
			
			// ========== FIX DUPLICATE STUDENT IDs ==========
			
			// Save original StudentID
			original := sid
			
			// While this StudentID already exists
			// seen[sid] = check if sid exists in map
			for seen[sid] {
				// Make it unique by adding user ID
				sid = fmt.Sprintf("%s_%d", original, u.ID)
				// Example: "alice" → "alice_5"
				needsUpdate = true
			}
			
			// Mark this StudentID as used
			// seen["alice"] = true
			seen[sid] = true
			
			// ========== UPDATE USER IF CHANGED ==========
			
			// If StudentID changed, update it
			if u.StudentID != sid {
				u.StudentID = sid
				needsUpdate = true
			}
			
			// If Phone is empty, set default
			if strings.TrimSpace(u.Phone) == "" {
				u.Phone = "0000000000"
				needsUpdate = true
			}
			
			// If anything changed, save to database
			if needsUpdate {
				database.DB.Save(&u)
			}
		}
	}
	
	// ========== PART 2: FIX ITEM DATA ==========
	
	// Set approval_status to "approved" for items that don't have it
	// Model(&model.Item{}) = work with items table
	// Where(...) = filter rows
	// Update("approval_status", "approved") = set column to value
	// SQL: UPDATE items SET approval_status = 'approved' 
	//      WHERE approval_status IS NULL OR approval_status = ''
	database.DB.Model(&model.Item{}).
		Where("approval_status IS NULL OR approval_status = ''").
		Update("approval_status", "approved")
	
	// ========== PART 3: FIX CLAIM DATA ==========
	
	// Set request_type to "claim_request" for claims that don't have it
	// SQL: UPDATE claims SET request_type = 'claim_request' 
	//      WHERE request_type IS NULL OR request_type = ''
	database.DB.Model(&model.Claim{}).
		Where("request_type IS NULL OR request_type = ''").
		Update("request_type", "claim_request")
}

// ========== HELPER FUNCTION: ADD DATABASE CONSTRAINTS ==========
func enforceUserConstraints() {
	// Execute raw SQL commands
	// Exec = execute SQL without returning rows
	
	// Drop old index if it exists
	// INDEX = database index for faster queries
	database.DB.Exec("DROP INDEX IF EXISTS idx_users_student_id")
	
	// Create unique index on lowercase student_id
	// UNIQUE = no two users can have same student_id
	// LOWER(student_id) = case-insensitive uniqueness
	database.DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_users_student_id_lower ON users (LOWER(student_id))")
	
	// Make student_id column required (NOT NULL)
	// ALTER TABLE = modify table structure
	// SET NOT NULL = column cannot be empty
	database.DB.Exec("ALTER TABLE users ALTER COLUMN student_id SET NOT NULL")
	
	// Make phone column required (NOT NULL)
	database.DB.Exec("ALTER TABLE users ALTER COLUMN phone SET NOT NULL")
}
```

---

# 2️⃣ INTERNAL/HANDLER/AUTH_HANDLER.GO - Authentication Handler

```go
// Package declaration
package handler

// Import required packages
import (
	"lostfound/internal/middleware"  // For session management
	"lostfound/internal/service"     // For business logic
	"net/http"                       // For HTTP status codes
	"github.com/gin-gonic/gin"       // Web framework
)

// ========== STRUCT DEFINITION ==========

// AuthHandler struct holds authentication-related methods
// struct = custom data type that groups related data
type AuthHandler struct {
	// authService = pointer to AuthService
	// *service.AuthService = pointer type
	authService *service.AuthService
}

// ========== CONSTRUCTOR FUNCTION ==========

// NewAuthHandler creates and returns a new AuthHandler instance
// *AuthHandler = returns pointer to AuthHandler
func NewAuthHandler() *AuthHandler {
	// &AuthHandler{} = create new AuthHandler and return its address
	return &AuthHandler{
		// Initialize authService field
		authService: service.NewAuthService(),
	}
}

// ========== SHOW LOGIN PAGE ==========

// ShowLogin displays the login page
// (h *AuthHandler) = method belongs to AuthHandler
// h = receiver (like "this" in other languages)
func (h *AuthHandler) ShowLogin(c *gin.Context) {
	// Render HTML template
	// renderHTML = helper function (defined in render.go)
	// http.StatusOK = 200 status code
	// "login.html" = template file name
	// gin.H = map[string]interface{} shortcut
	renderHTML(c, http.StatusOK, "login.html", gin.H{
		"title":            "Login",         // Page title
		"content_template": "login_content", // Sub-template name
	})
}

// ========== SHOW REGISTRATION PAGE ==========

// ShowRegister displays the registration page
func (h *AuthHandler) ShowRegister(c *gin.Context) {
	// Same as ShowLogin but for registration
	renderHTML(c, http.StatusOK, "register.html", gin.H{
		"title":            "Register",
		"content_template": "register_content",
	})
}

// ========== PROCESS LOGIN FORM ==========

// Login processes the login form submission
func (h *AuthHandler) Login(c *gin.Context) {
	// ========== STEP 1: GET FORM DATA ==========
	
	// Get student_id from POST form
	// c.PostForm("student_id") = get value of form field named "student_id"
	studentID := c.PostForm("student_id")
	
	// Get password from POST form
	password := c.PostForm("password")

	// ========== STEP 2: VERIFY CREDENTIALS ==========
	
	// Call authService.Login to check credentials
	// Returns: (user, error)
	// user = User object if successful
	// err = error message if failed
	user, err := h.authService.Login(studentID, password)
	
	// ========== STEP 3: HANDLE LOGIN FAILURE ==========
	
	// If error occurred (invalid credentials)
	if err != nil {
		// Show login page again with error message
		renderHTML(c, http.StatusOK, "login.html", gin.H{
			"title":            "Login",
			"error":            err.Error(),  // Convert error to string
			"content_template": "login_content",
		})
		return  // Stop execution
	}

	// ========== STEP 4: CREATE SESSION ==========
	
	// Login successful, create session
	
	// Get session from middleware
	// session = cookie-based session storage
	session := middleware.GetSession(c)
	
	// Store user ID in session
	// session.Values = map to store session data
	// This creates a cookie in the browser
	session.Values["user_id"] = user.ID
	
	// Save session (writes cookie to browser)
	// c.Request = HTTP request object
	// c.Writer = HTTP response writer
	session.Save(c.Request, c.Writer)

	// ========== STEP 5: REDIRECT BASED ON ROLE ==========
	
	// If user is admin, redirect to admin dashboard
	if user.Role == "admin" {
		// http.StatusSeeOther = 303 redirect (POST → GET)
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	} else {
		// Regular user, redirect to user dashboard
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// ========== PROCESS REGISTRATION FORM ==========

// Register processes the registration form submission
func (h *AuthHandler) Register(c *gin.Context) {
	// ========== STEP 1: GET FORM DATA ==========
	
	// Get all form fields
	name := c.PostForm("name")
	studentID := c.PostForm("student_id")
	phone := c.PostForm("phone")
	password := c.PostForm("password")

	// ========== STEP 2: CREATE USER ACCOUNT ==========
	
	// Call authService.Register to create new user
	// This validates data, hashes password, saves to database
	user, err := h.authService.Register(name, studentID, phone, password)
	
	// ========== STEP 3: HANDLE REGISTRATION FAILURE ==========
	
	// If error occurred (e.g., student ID already exists)
	if err != nil {
		// Show registration page again with error
		renderHTML(c, http.StatusOK, "register.html", gin.H{
			"title":            "Register",
			"error":            err.Error(),
			"content_template": "register_content",
		})
		return
	}

	// ========== STEP 4: AUTO-LOGIN AFTER REGISTRATION ==========
	
	// Registration successful, automatically log in the user
	
	// Get session
	session := middleware.GetSession(c)
	
	// Store user ID in session
	session.Values["user_id"] = user.ID
	
	// Save session
	session.Save(c.Request, c.Writer)

	// ========== STEP 5: REDIRECT TO DASHBOARD ==========
	
	// Redirect to user dashboard
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

// ========== LOGOUT USER ==========

// Logout logs out the current user
func (h *AuthHandler) Logout(c *gin.Context) {
	// ========== STEP 1: GET SESSION ==========
	
	// Get current session
	session := middleware.GetSession(c)
	
	// ========== STEP 2: CLEAR SESSION DATA ==========
	
	// Create empty map to replace session data
	// make(map[interface{}]interface{}) = create empty map
	// This effectively deletes all session data (including user_id)
	session.Values = make(map[interface{}]interface{})
	
	// ========== STEP 3: SAVE EMPTY SESSION ==========
	
	// Save empty session (deletes the cookie)
	session.Save(c.Request, c.Writer)
	
	// ========== STEP 4: REDIRECT TO HOME ==========
	
	// Redirect to home page
	c.Redirect(http.StatusSeeOther, "/")
}
```

---

*Due to length constraints, I'll create this as a file. The document continues with all other files...*

