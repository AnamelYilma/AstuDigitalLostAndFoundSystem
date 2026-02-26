package handler

import (
	"lostfound/internal/model"
	"net/http"
	"strconv"
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
	user, _ := c.Get("user")
	stats, _ := h.itemService.GetStats()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":            "Dashboard",
		"user":             user,
		"stats":            stats,
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
	brand := c.PostForm("brand")
	location := c.PostForm("location")
	date := c.PostForm("date")
	description := c.PostForm("description")

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
			"content_template": "report_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ItemHandler) ShowSearch(c *gin.Context) {
	c.HTML(http.StatusOK, "search.html", gin.H{
		"title":            "Search Items",
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
	if itemType := c.Query("type"); itemType != "" {
		filters["type"] = itemType
	}

	items, err := h.itemService.SearchItems(filters)
	if err != nil {
		items = []model.Item{}
	}

	c.HTML(http.StatusOK, "items.html", gin.H{
		"title":            "Search Results",
		"items":            items,
		"filters":          filters,
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
