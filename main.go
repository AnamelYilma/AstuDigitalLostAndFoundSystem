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

