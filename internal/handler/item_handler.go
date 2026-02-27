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

	renderHTML(c, http.StatusOK, "report.html", gin.H{
		"title":            "Report Item",
		"user":             user,
		"type":             itemType,
		"locations":        service.ASTULocations(),
		"colors":           service.ColorOptions(),
		"content_template": "report_content",
	})
}

func (h *ItemHandler) ReportItem(c *gin.Context) {
	u := c.MustGet("user").(model.User)

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
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
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
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
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
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
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
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}

	file, err := c.FormFile("image")
	imagePath := ""
	if err != nil && itemType == "found" {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
			"type":             itemType,
			"error":            "Photo is required for found item reports",
			"locations":        service.ASTULocations(),
			"colors":           service.ColorOptions(),
			"content_template": "report_content",
		})
		return
	}
	if err == nil {
		imagePath, err = h.itemService.SaveImage(file)
		if err != nil {
			renderHTML(c, http.StatusOK, "report.html", gin.H{
				"title":            "Report Item",
				"user":             u,
				"type":             itemType,
				"error":            "Image upload failed: " + err.Error(),
				"locations":        service.ASTULocations(),
				"colors":           service.ColorOptions(),
				"content_template": "report_content",
			})
			return
		}
	}

	_, err = h.itemService.CreateItem(
		u.ID, itemType, title, category, color, brand,
		location, date, description, imagePath,
	)

	if err != nil {
		renderHTML(c, http.StatusOK, "report.html", gin.H{
			"title":            "Report Item",
			"user":             u,
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

	if category := c.Query("category"); category != "" {
		if service.IsValidCategory(category) {
			filters["category"] = category
		}
	}
	if location := c.Query("location"); location != "" {
		if service.IsValidASTULocation(location) {
			filters["location"] = location
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
	}
	if itemType := c.Query("type"); itemType != "" {
		if service.IsValidItemType(itemType) {
			filters["type"] = itemType
		}
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = dateFrom
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = dateTo
		}
	}
	if status := c.Query("status"); status != "" {
		if service.IsValidApprovalStatus(status) {
			// Public explore only shows approved items.
			if isAdmin {
				filters["approval_status"] = status
			} else if status == "approved" {
				filters["approval_status"] = "approved"
			}
		}
	}
	if !isAdmin {
		if _, ok := filters["approval_status"]; !ok {
			filters["approval_status"] = "approved"
		}
	}

	items, err := h.itemService.SearchItems(filters)
	if err != nil {
		items = []model.Item{}
	}

	renderHTML(c, http.StatusOK, "items.html", gin.H{
		"title":            "Search Results",
		"items":            items,
		"filters":          filters,
		"selected_colors":  selectedColors,
		"user":             user,
		"is_admin":         isAdmin,
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
			c.Redirect(http.StatusSeeOther, "/search")
			return
		}
		u := user.(model.User)
		if u.Role != "admin" && u.ID != item.UserID {
			c.Redirect(http.StatusSeeOther, "/search")
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
