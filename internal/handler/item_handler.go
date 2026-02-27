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
	user := c.MustGet("user").(model.User)
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
	user := c.MustGet("user").(model.User)
	itemType := c.Query("type")
	if itemType != "lost" && itemType != "found" {
		itemType = "lost"
	}

	c.HTML(http.StatusOK, "report.html", gin.H{
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
	if itemType != "lost" && itemType != "found" {
		c.HTML(http.StatusOK, "report.html", gin.H{
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
	if !service.IsValidASTULocation(location) {
		c.HTML(http.StatusOK, "report.html", gin.H{
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
		c.HTML(http.StatusOK, "report.html", gin.H{
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
