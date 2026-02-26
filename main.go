package main

import (
	"html/template"
	"log"
	"lostfound/internal/handler"
	"lostfound/internal/middleware"
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"lostfound/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database.InitDB()

	// Auto migrate tables
	database.DB.AutoMigrate(&model.User{}, &model.Item{}, &model.Claim{})

	// Create default admin if not exists
	createDefaultAdmin()

	// Setup Gin
	r := gin.Default()

	// Add template function
	r.SetFuncMap(template.FuncMap{
		"now": func() string {
			return time.Now().Format("2006-01-02")
		},
	})

	// FIXED: Load templates properly - layout FIRST, then pages
	r.LoadHTMLFiles(
		"templates/layout.html",
		"templates/index.html",
		"templates/login.html",
		"templates/register.html",
		"templates/error.html",
		"templates/dashboard.html",
		"templates/report.html",
		"templates/search.html",
		"templates/items.html",
		"templates/item.html",
		"templates/admin/dashboard.html",
		"templates/admin/claims.html",
	)

	// Serve static files
	r.Static("/static", "./static")

	// Apply middleware
	r.Use(middleware.SetUser())

	// Initialize handlers
	authHandler := handler.NewAuthHandler()
	itemHandler := handler.NewItemHandler()
	adminHandler := handler.NewAdminHandler()

	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title":            "ASTU Lost & Found",
			"content_template": "index_content",
		})
	})

	// Auth routes
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	r.GET("/register", authHandler.ShowRegister)
	r.POST("/register", authHandler.Register)
	r.GET("/logout", authHandler.Logout)

	// Search (public)
	r.GET("/search", itemHandler.ShowSearch)
	r.GET("/items", itemHandler.Search)
	r.GET("/item/:id", itemHandler.ShowItem)

	// Protected routes (require login)
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/dashboard", itemHandler.Dashboard)
		protected.GET("/report", itemHandler.ShowReportForm)
		protected.POST("/report", itemHandler.ReportItem)
		protected.POST("/claim", itemHandler.ClaimItem)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/dashboard", adminHandler.Dashboard)
		admin.GET("/claims", adminHandler.ShowClaims)
		admin.POST("/admin/claims/update", adminHandler.UpdateClaim)
	}

	// Start server
	log.Println("✅ Server starting on http://localhost:8080")
	r.Run(":8080")
}

func createDefaultAdmin() {
	var count int64
	database.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

	if count == 0 {
		// Create default admin
		hashedPassword, _ := utils.HashPassword("admin123")
		admin := &model.User{
			Name:     "Admin",
			Email:    "admin@astu.edu",
			Password: hashedPassword,
			Role:     "admin",
		}
		database.DB.Create(admin)
		log.Println("✅ Default admin created: admin@astu.edu / admin123")
	}
}
