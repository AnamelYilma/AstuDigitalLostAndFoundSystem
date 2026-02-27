package main

import (
	"fmt"
	"html/template"
	"log"
	"lostfound/internal/handler"
	"lostfound/internal/middleware"
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"lostfound/pkg/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	database.DB.AutoMigrate(&model.User{}, &model.Item{}, &model.Claim{})
	createDefaultAdmin()
	normalizeLegacyData()

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"now": func() string {
			return time.Now().Format("2006-01-02")
		},
	})

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
		"templates/admin/admin_dashboard.html",
		"templates/admin/admin_claims.html",
		"templates/admin/admin_items.html",
	)

	r.Static("/static", "./static")
	r.Use(middleware.SetUser())

	authHandler := handler.NewAuthHandler()
	itemHandler := handler.NewItemHandler()
	adminHandler := handler.NewAdminHandler()

	r.GET("/", func(c *gin.Context) {
		user, _ := c.Get("user")
		c.HTML(200, "index.html", gin.H{
			"title":            "ASTU Lost & Found",
			"user":             user,
			"content_template": "index_content",
		})
	})

	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	r.GET("/register", authHandler.ShowRegister)
	r.POST("/register", authHandler.Register)
	r.GET("/logout", authHandler.Logout)

	r.GET("/search", itemHandler.ShowSearch)
	r.GET("/items", itemHandler.Search)
	r.GET("/item/:id", itemHandler.ShowItem)

	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/dashboard", itemHandler.Dashboard)
		protected.GET("/report", itemHandler.ShowReportForm)
		protected.POST("/report", itemHandler.ReportItem)
		protected.POST("/claim", itemHandler.ClaimItem)
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

	log.Println("Server starting on http://localhost:8080")
	r.Run(":8080")
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
	if err := database.DB.Where("student_id IS NULL OR student_id = '' OR phone IS NULL OR phone = ''").Find(&users).Error; err == nil {
		for _, u := range users {
			needsUpdate := false
			if strings.TrimSpace(u.StudentID) == "" {
				baseID := strings.Split(strings.TrimSpace(u.Email), "@")[0]
				if baseID == "" {
					baseID = fmt.Sprintf("user_%d", u.ID)
				}
				u.StudentID = strings.ToLower(baseID)
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
}
