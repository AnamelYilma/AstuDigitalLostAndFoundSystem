This is my project Folder structure :
```

├── 📁 internal
│   ├── 📁 handler
│   │   ├── 🐹 admin_handler.go
│   │   ├── 🐹 auth_handler.go
│   │   └── 🐹 item_handler.go
│   ├── 📁 middleware
│   │   └── 🐹 auth_middleware.go
│   ├── 📁 model
│   │   ├── 🐹 item.go
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
│   └── 📁 js
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
│   ├── 🌐 register.html
│   ├── 🌐 report.html
│   └── 🌐 search.html
├── ⚙️ .gitignore
├── 📝 ASTU STEAM COMMAND.md
├── 📄 go.mod
├── 📄 go.sum
└── 🐹 main.go
```




addmin handerl go"package handler

import (
	"lostfound/internal/service"
	"net/http"
	"strconv"

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

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
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

	c.HTML(http.StatusOK, "admin_claims.html", gin.H{
		"title":            "Manage Claims",
		"user":             user,
		"claims":           claims,
		"content_template": "admin_claims_content",
	})
}

func (h *AdminHandler) ShowItems(c *gin.Context) {
	user, _ := c.Get("user")
	items, _ := h.itemService.GetAllItemsForAdmin()

	c.HTML(http.StatusOK, "admin_items.html", gin.H{
		"title":            "Manage Item Posts",
		"user":             user,
		"items":            items,
		"content_template": "admin_items_content",
	})
}

func (h *AdminHandler) UpdateClaim(c *gin.Context) {
	claimID, _ := strconv.ParseUint(c.PostForm("claim_id"), 10, 32)
	status := c.PostForm("status")
	remarks := c.PostForm("remarks")

	err := h.itemService.UpdateClaimStatus(uint(claimID), status, remarks)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{
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
		c.HTML(http.StatusOK, "error.html", gin.H{
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
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error":            "Failed to remove item: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/items")
}
"

auth handler go

"
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
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title":            "Login",
		"content_template": "login_content",
	})
}

func (h *AuthHandler) ShowAdminLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_login.html", gin.H{
		"title":            "Admin Login",
		"content_template": "admin_login_content",
	})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":            "Register",
		"content_template": "register_content",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	studentID := c.PostForm("student_id")
	password := c.PostForm("password")

	user, err := h.authService.Login(studentID, password)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":            "Login",
			"error":            err.Error(),
			"content_template": "login_content",
		})
		return
	}

	if user.Role == "admin" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":            "Login",
			"error":            "Admin must use Admin Login page",
			"content_template": "login_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	studentID := c.PostForm("student_id")
	password := c.PostForm("password")

	user, err := h.authService.Login(studentID, password)
	if err != nil || user.Role != "admin" {
		c.HTML(http.StatusOK, "admin_login.html", gin.H{
			"title":            "Admin Login",
			"error":            "invalid admin ID or password",
			"content_template": "admin_login_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusSeeOther, "/admin/dashboard")
}

func (h *AuthHandler) Register(c *gin.Context) {
	name := c.PostForm("name")
	studentID := c.PostForm("student_id")
	phone := c.PostForm("phone")
	password := c.PostForm("password")

	user, err := h.authService.Register(name, studentID, phone, password)
	if err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
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

"
item hanlder go"
package handler

import (
	"lostfound/internal/model"
	"net/http"
	"strconv"
	"strings"
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
	userAny, _ := c.Get("user")
	user := userAny.(model.User)
	stats, _ := h.itemService.GetStats()
	myItems, _ := h.itemService.GetItemsByUserID(user.ID)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":            "Dashboard",
		"user":             user,
		"stats":            stats,
		"my_items":         myItems,
		"content_template": "dashboard_content",
	})
}

func (h *ItemHandler) ShowReportForm(c *gin.Context) {
	itemType := c.Query("type")
	if itemType != "lost" && itemType != "found" {
		itemType = "lost"
	}

	c.HTML(http.StatusOK, "report.html", gin.H{
		"title":            "Report Item",
		"type":             itemType,
		"locations":        service.ASTULocations(),
		"colors":           service.ColorOptions(),
		"content_template": "report_content",
	})
}

func (h *ItemHandler) ReportItem(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(model.User)

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
	if itemType != "lost" && itemType != "found" {
		c.HTML(http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"type":             "lost",
			"error":            "Invalid report type",
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}
	if !service.IsValidASTULocation(location) {
		c.HTML(http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"type":             itemType,
			"error":            "Please select a valid ASTU location",
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}
	if strings.TrimSpace(color) == "" {
		c.HTML(http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"type":             itemType,
			"error":            "Please provide item color",
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}

	file, err := c.FormFile("image")
	imagePath := ""
	if err == nil {
		imagePath, _ = h.itemService.SaveImage(file)
	}

	_, err = h.itemService.CreateItem(
		u.ID, itemType, title, category, color, brand,
		location, date, description, imagePath,
	)

	if err != nil {
		c.HTML(http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"type":             itemType,
			"error":            "Failed to save item: " + err.Error(),
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ItemHandler) ShowSearch(c *gin.Context) {
	user, _ := c.Get("user")
	c.HTML(http.StatusOK, "search.html", gin.H{
		"title":            "Search Items",
		"user":             user,
		"locations":        service.ASTULocations(),
		"colors":           service.ColorOptions(),
		"content_template": "search_content",
	})
}

func (h *ItemHandler) Search(c *gin.Context) {
	filters := make(map[string]interface{})

	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
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
	}
	if itemType := c.Query("type"); itemType != "" {
		if itemType == "lost" || itemType == "found" {
			filters["type"] = itemType
		}
	}
	if status := c.Query("status"); status != "" {
		if status == "pending" || status == "approved" || status == "rejected" {
			filters["approval_status"] = status
		}
	}

	items, err := h.itemService.SearchItems(filters)
	if err != nil {
		items = []model.Item{}
	}

	user, _ := c.Get("user")
	c.HTML(http.StatusOK, "items.html", gin.H{
		"title":            "Search Results",
		"items":            items,
		"filters":          filters,
		"selected_colors":  selectedColors,
		"user":             user,
		"locations":        service.ASTULocations(),
		"colors":           service.ColorOptions(),
		"content_template": "items_content",
	})
}

func (h *ItemHandler) ShowItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	item, err := h.itemService.GetItemByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/search")
		return
	}

	user, _ := c.Get("user")

	c.HTML(http.StatusOK, "item.html", gin.H{
		"title":            item.Title,
		"item":             item,
		"user":             user,
		"content_template": "item_content",
	})
}

func (h *ItemHandler) ClaimItem(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(model.User)

	itemID, _ := strconv.ParseUint(c.PostForm("item_id"), 10, 32)
	description := c.PostForm("description")

	err := h.itemService.CreateClaim(uint(itemID), u.ID, description)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error":            "Failed to submit claim: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

"
auth middleware go "
package middleware

import (
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("your-secret-key-change-this"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
	}
}

func GetSession(c *gin.Context) *sessions.Session {
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

"
item go "
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

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Claim struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ItemID       uint           `gorm:"not null" json:"item_id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
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

"
user go "
package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	StudentID string         `gorm:"size:30;index" json:"student_id"`
	Phone     string         `gorm:"size:20" json:"phone"`
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

"
item repotory go"
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
	query := database.DB.Preload("User")

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
	if approvalStatus, ok := filters["approval_status"]; ok && approvalStatus != "" {
		query = query.Where("approval_status = ?", approvalStatus)
	}

	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindByID(id uint) (*model.Item, error) {
	var item model.Item
	err := database.DB.Preload("User").First(&item, id).Error
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

func (r *ItemRepository) FindByUserID(userID uint) ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindAllItemsForAdmin() ([]model.Item, error) {
	var items []model.Item
	err := database.DB.Preload("User").Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *ItemRepository) DeleteItem(itemID uint) error {
	return database.DB.Delete(&model.Item{}, itemID).Error
}

"

user repotory go"
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

"
auth service go "package service

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
"
item option go "
package service

import "strings"

var astuLocations = []string{
	"Library",
	"Cafe",
	"Class",
	"Lap",
	"Dorm",
	"On Road",
	"Tolest",
	"Shower",
	"Anphe",
	"Launch",
	"Park",
	"Hale.Birroe",
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

func ASTULocations() []string {
	return astuLocations
}

func ColorOptions() []string {
	return colorOptions
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

func IsValidASTULocation(location string) bool {
	for _, l := range astuLocations {
		if strings.EqualFold(strings.TrimSpace(location), l) {
			return true
		}
	}
	return false
}

"
item serviece go "
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

func (s *ItemService) CreateItem(userID uint, itemType, title, category, color, brand, location, date, description, imagePath string) (*model.Item, error) {
	if itemType != "lost" && itemType != "found" {
		return nil, errors.New("invalid item type")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("item title is required")
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
		Image:          imagePath,
		Status:         "open",
		ApprovalStatus: "pending",
	}

	err := s.itemRepo.Create(item)
	return item, err
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
	if item.Type != "found" {
		return errors.New("you can only claim found items")
	}
	if item.Status != "open" {
		return errors.New("this item is no longer available for claim")
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("claim description is required")
	}

	claim := &model.Claim{
		ItemID:      itemID,
		UserID:      userID,
		Description: description,
		Status:      "pending",
	}

	return s.itemRepo.CreateClaim(claim)
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
	if status != "approved" && status != "rejected" && status != "pending" {
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
	return s.itemRepo.Update(item)
}

func (s *ItemService) DeleteItem(itemID uint) error {
	return s.itemRepo.DeleteItem(itemID)
}

"

db go "
package database

import (
    "fmt"
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
    dsn := "host=localhost user=postgres password=0909 dbname=lostfound port=5432 sslmode=disable"
    
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        log.Fatal("❌ Failed to connect to database:", err)
    }
    
    fmt.Println("✅ Database connected successfully")
}
"
hash go "
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
"
style.css "

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
    max-width: 1200px;
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
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #1a365d;
    box-shadow: 0 0 0 2px rgba(26, 54, 93, 0.1);
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
    
    .btn {
        width: 100%;
        text-align: center;
    }
}
"


admin login html "
<!-- templates/admin_login.html -->
{{ template "layout.html" . }}

{{ define "admin_login_content" }}
<div class="auth-form">
    <h2>Admin Login</h2>

    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}

    <form method="POST" action="/admin/login">
        <div class="form-group">
            <label>Admin ID</label>
            <input type="text" name="student_id" required placeholder="admin">
        </div>

        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required>
        </div>

        <button type="submit" class="btn btn-primary">Login as Admin</button>
    </form>
</div>
{{ end }}

"
dashboard html " 
<!-- templates/admin_login.html -->
{{ template "layout.html" . }}

{{ define "admin_login_content" }}
<div class="auth-form">
    <h2>Admin Login</h2>

    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}

    <form method="POST" action="/admin/login">
        <div class="form-group">
            <label>Admin ID</label>
            <input type="text" name="student_id" required placeholder="admin">
        </div>

        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required>
        </div>

        <button type="submit" class="btn btn-primary">Login as Admin</button>
    </form>
</div>
{{ end }}

"
error html "
<!-- templates/error.html -->
{{ template "layout.html" . }}

{{ define "error_content" }}
<div class="error-page">
    <h2>{{ .title }}</h2>
    <div class="error">{{ .error }}</div>
    <a href="/" class="btn btn-secondary">Back Home</a>
</div>
{{ end }}

"
index html "
 <!-- templates/index.html -->
{{ template "layout.html" . }}

{{ define "index_content" }}
<div class="hero">
    <h1>ASTU Digital Lost & Found</h1>
    <p>Report lost items, find missing belongings, and reunite with your possessions</p>
    
    <div class="hero-buttons">
        <a href="/search" class="btn btn-primary">Search Items</a>
        <a href="/report?type=found" class="btn btn-secondary">Report Found Item</a>
        <a href="/report?type=lost" class="btn btn-secondary">Report Lost Item</a>
    </div>
</div>

<div class="features">
    <div class="feature-card">
        <h3>Report Lost Items</h3>
        <p>Lost something? Report it here and increase your chances of finding it.</p>
    </div>
    <div class="feature-card">
        <h3>Report Found Items</h3>
        <p>Found something? Help someone get their belonging back.</p>
    </div>
    <div class="feature-card">
        <h3>Search & Claim</h3>
        <p>Search for your lost items and claim them when found.</p>
    </div>
</div>
{{ end }}

"
item html "
<!-- templates/item.html -->
{{ template "layout.html" . }}

{{ define "item_content" }}
<div class="item-detail">
    <div class="item-header">
        <h2>{{ .item.Title }}</h2>
        <span class="badge {{ .item.Type }}">{{ .item.Type }}</span>
    </div>
    
    <div class="item-content">
        {{ if .item.Image }}
        <div class="item-image-large">
            <img src="{{ .item.Image }}" alt="{{ .item.Title }}">
        </div>
        {{ end }}
        
        <div class="item-info">
            <p><strong>Category:</strong> {{ .item.Category }}</p>
            <p><strong>Color:</strong> {{ .item.Color }}</p>
            <p><strong>Brand:</strong> {{ .item.Brand }}</p>
            <p><strong>Location:</strong> {{ .item.Location }}</p>
            <p><strong>Date:</strong> {{ .item.Date }}</p>
            <p><strong>Reported by:</strong> {{ .item.User.Name }}</p>
            <p><strong>Contact:</strong> {{ .item.User.StudentID }} / {{ .item.User.Phone }}</p>
            <p><strong>Approval:</strong> {{ .item.ApprovalStatus }}</p>
            <p><strong>Description:</strong></p>
            <p>{{ .item.Description }}</p>
        </div>
    </div>
    
    {{ if and (eq .item.Type "found") .user (ne .user.ID .item.UserID) (eq .item.Status "open") (eq .item.ApprovalStatus "approved") }}
    <div class="claim-section">
        <h3>Claim This Item</h3>
        <form method="POST" action="/claim">
            <input type="hidden" name="item_id" value="{{ .item.ID }}">
            <div class="form-group">
                <label>Why do you think this is yours?</label>
                <textarea name="description" required rows="3"></textarea>
            </div>
            <button type="submit" class="btn btn-primary">Submit Claim</button>
        </form>
    </div>
    {{ end }}

    {{ if and (eq .item.Type "found") (not .user) }}
    <div class="claim-section">
        <h3>Want to claim this item?</h3>
        <p>Please login first, then submit your claim.</p>
        <a href="/login" class="btn btn-primary">Login to Claim</a>
    </div>
    {{ end }}

    {{ if eq .item.Type "lost" }}
    <div class="claim-section">
        <h3>Did you find this item?</h3>
        <a href="/report?type=found" class="btn btn-primary">Report Found Item</a>
    </div>
    {{ end }}
    
    <a href="/search" class="btn btn-secondary">Back to Search</a>
</div>
{{ end }}

"
items html "
<!-- templates/items.html -->
{{ template "layout.html" . }}

{{ define "items_content" }}
<div class="items-list">
    <h2>Search Results</h2>
    <p>
        <strong>Filters:</strong>
        Type: {{ if index .filters "type" }}{{ index .filters "type" }}{{ else }}all{{ end }} |
        Category: {{ if index .filters "category" }}{{ index .filters "category" }}{{ else }}all{{ end }} |
        Location: {{ if index .filters "location" }}{{ index .filters "location" }}{{ else }}all{{ end }} |
        Color: {{ if .selected_colors }}{{ range $i, $c := .selected_colors }}{{ if $i }}, {{ end }}{{ $c }}{{ end }}{{ else }}all{{ end }} |
        Status: {{ if index .filters "approval_status" }}{{ index .filters "approval_status" }}{{ else }}all{{ end }}
    </p>
    
    {{ if .items }}
    <div class="items-grid">
        {{ range .items }}
        <div class="item-card">
            {{ if .Image }}
            <img src="{{ .Image }}" alt="{{ .Title }}" class="item-image">
            {{ else }}
            <div class="no-image">No Image</div>
            {{ end }}
            
            <div class="item-details">
                <h3>{{ .Title }}</h3>
                <p><strong>Type:</strong> {{ .Type }}</p>
                <p><strong>Category:</strong> {{ .Category }}</p>
                <p><strong>Location:</strong> {{ .Location }}</p>
                <p><strong>Date:</strong> {{ .Date }}</p>
                <p><strong>Status:</strong> {{ .Status }}</p>
                <p><strong>Approval:</strong> {{ .ApprovalStatus }}</p>
                <a href="/item/{{ .ID }}" class="btn btn-small">View Details</a>
            </div>
        </div>
        {{ end }}
    </div>
    {{ else }}
    <p>No items found matching your criteria.</p>
    {{ end }}
</div>
{{ end }}

"
layout html "
<!-- templates/layout.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .title }} - ASTU Lost & Found</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="container">
            <a href="/" class="logo">ASTU Lost & Found</a>
            <div class="nav-links">
                <a href="/">Home</a>
                <a href="/search">Search</a>
                {{ if .user }}
                    <a href="/report?type=lost">Report Lost</a>
                    <a href="/report?type=found">Report Found</a>
                    {{ if eq .user.Role "admin" }}
                        <a href="/admin/dashboard">Admin</a>
                    {{ end }}
                    <a href="/logout">Logout ({{ .user.Name }})</a>
                {{ else }}
                    <a href="/login">Login</a>
                    <a href="/admin/login">Admin Login</a>
                    <a href="/register">Register</a>
                {{ end }}
            </div>
        </div>
    </nav>
    
    <main class="container">
        {{ if eq .content_template "index_content" }}
            {{ template "index_content" . }}
        {{ else if eq .content_template "login_content" }}
            {{ template "login_content" . }}
        {{ else if eq .content_template "admin_login_content" }}
            {{ template "admin_login_content" . }}
        {{ else if eq .content_template "register_content" }}
            {{ template "register_content" . }}
        {{ else if eq .content_template "dashboard_content" }}
            {{ template "dashboard_content" . }}
        {{ else if eq .content_template "report_content" }}
            {{ template "report_content" . }}
        {{ else if eq .content_template "search_content" }}
            {{ template "search_content" . }}
        {{ else if eq .content_template "items_content" }}
            {{ template "items_content" . }}
        {{ else if eq .content_template "item_content" }}
            {{ template "item_content" . }}
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
    
    <footer class="footer">
        <div class="container">
            <p>&copy; 2024 ASTU Lost & Found System</p>
        </div>
    </footer>
</body>
</html>

"
login html "
<!-- templates/login.html -->
{{ template "layout.html" . }}

{{ define "login_content" }}
<div class="auth-form">
    <h2>Login</h2>
    
    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}
    
    <form method="POST" action="/login">
        <div class="form-group">
            <label>Student ID</label>
            <input type="text" name="student_id" required placeholder="ugr/38923/18">
        </div>
        
        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required>
        </div>
        
        <button type="submit" class="btn btn-primary">Login</button>
    </form>
    
    <p class="auth-link">Don't have an account? <a href="/register">Register</a></p>
    <p class="auth-link">Admin? <a href="/admin/login">Use admin login</a></p>
    
    <div class="demo-credentials">
        <p><strong>Demo Credentials:</strong></p>
        <p>Admin ID: admin / admin123</p>
    </div>
</div>
{{ end }}

"
registor html "
<!-- templates/register.html -->
{{ template "layout.html" . }}

{{ define "register_content" }}
<div class="auth-form">
    <h2>Register</h2>

    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}

    <form method="POST" action="/register">
        <div class="form-group">
            <label>Full Name</label>
            <input type="text" name="name" required placeholder="John Doe">
        </div>

        <div class="form-group">
            <label>Student ID</label>
            <input type="text" name="student_id" required placeholder="ugr/38923/18">
        </div>

        <div class="form-group">
            <label>Phone Number</label>
            <input type="text" name="phone" required placeholder="09xxxxxxxx">
        </div>

        <div class="form-group">
            <label>Password</label>
            <input type="password" name="password" required minlength="6">
        </div>

        <button type="submit" class="btn btn-primary">Register</button>
    </form>

    <p class="auth-link">Already have an account? <a href="/login">Login</a></p>
</div>
{{ end }}

"

report html 
"
<!-- templates/report.html -->
{{ template "layout.html" . }}

{{ define "report_content" }}
<div class="report-form">
    <h2>Report {{ if eq .type "lost" }}Lost{{ else }}Found{{ end }} Item</h2>
    
    {{ if .error }}
    <div class="error">{{ .error }}</div>
    {{ end }}
    
    <form method="POST" action="/report" enctype="multipart/form-data">
        <input type="hidden" name="type" value="{{ .type }}">
        
        <div class="form-group">
            <label>Category *</label>
            <select name="category" required>
                <option value="electronics">Electronics</option>
                <option value="id">ID Card</option>
                <option value="books">Books</option>
                <option value="clothing">Clothing</option>
                <option value="accessories">Accessories</option>
                <option value="other">Other</option>
            </select>
        </div>
        
        <div class="form-group">
            <label>Item Name *</label>
            <input type="text" name="title" required placeholder="e.g., HP Calculator">
        </div>
        
        <div class="form-row">
            <div class="form-group">
                <label>Color</label>
                <select name="color">
                    {{ range .colors }}
                    <option value="{{ . }}">{{ . }}</option>
                    {{ end }}
                </select>
                <input type="text" name="color_other" placeholder="If other, type color name">
            </div>
            
            <div class="form-group">
                <label>Brand</label>
                <input type="text" name="brand" placeholder="e.g., HP">
            </div>
        </div>
        
        <div class="form-group">
            <label>Location *</label>
            <select name="location" required>
                {{ range .locations }}
                <option value="{{ . }}">{{ . }}</option>
                {{ end }}
            </select>
        </div>
        
        <div class="form-group">
            <label>Date</label>
            <input type="date" name="date" value="{{ now }}">
        </div>
        
        <div class="form-group">
            <label>Description</label>
            <textarea name="description" rows="4" placeholder="Describe your item..."></textarea>
        </div>
        
        <div class="form-group">
            <label>Photo</label>
            <input type="file" name="image" accept="image/*">
            <small>Max size: 5MB (JPG, PNG only)</small>
        </div>
        
        <button type="submit" class="btn btn-primary">Submit Report</button>
        <a href="/dashboard" class="btn btn-secondary">Cancel</a>
    </form>
</div>
{{ end }}

"
search html "
<!-- templates/search.html -->
{{ template "layout.html" . }}

{{ define "search_content" }}
<div class="search-page">
    <h2>Search Items</h2>
    
    <form method="GET" action="/items" class="search-form">
        <div class="form-row">
            <div class="form-group">
                <label>Category</label>
                <select name="category">
                    <option value="">All Categories</option>
                    <option value="electronics">Electronics</option>
                    <option value="id">ID Card</option>
                    <option value="books">Books</option>
                    <option value="clothing">Clothing</option>
                    <option value="accessories">Accessories</option>
                    <option value="other">Other</option>
                </select>
            </div>
            
            <div class="form-group">
                <label>Location</label>
                <select name="location">
                    <option value="">All Locations</option>
                    {{ range .locations }}
                    <option value="{{ . }}">{{ . }}</option>
                    {{ end }}
                </select>
            </div>

            <div class="form-group">
                <label>Color</label>
                <select name="color">
                    <option value="">All Colors</option>
                    {{ range .colors }}
                    <option value="{{ . }}">{{ . }}</option>
                    {{ end }}
                </select>
            </div>
            
            <div class="form-group">
                <label>Type</label>
                <select name="type">
                    <option value="">All Items</option>
                    <option value="lost">Lost Only</option>
                    <option value="found">Found Only</option>
                </select>
            </div>

            <div class="form-group">
                <label>Status</label>
                <select name="status">
                    <option value="">All Statuses</option>
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                </select>
            </div>
        </div>
        
        <button type="submit" class="btn btn-primary">Search</button>
        <a href="/items" class="btn btn-secondary">Show All</a>
    </form>
</div>
{{ end }}

"
main go "
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
		"templates/admin_login.html",
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
	r.GET("/admin/login", authHandler.ShowAdminLogin)
	r.POST("/admin/login", authHandler.AdminLogin)
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

"




