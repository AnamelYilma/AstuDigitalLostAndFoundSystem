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
