package handler

import (
	"lostfound/internal/model"
	"lostfound/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	renderHTML(c, http.StatusOK, "admin_dashboard.html", gin.H{
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

	renderHTML(c, http.StatusOK, "admin_claims.html", gin.H{
		"title":            "Manage Claims",
		"user":             user,
		"claims":           claims,
		"content_template": "admin_claims_content",
	})
}

func (h *AdminHandler) ShowItems(c *gin.Context) {
	user, _ := c.Get("user")
	filters := map[string]interface{}{}
	selectedQ := strings.TrimSpace(c.Query("q"))
	selectedType := ""
	selectedStatus := ""
	selectedCategory := ""
	selectedLocation := ""
	selectedDateFrom := ""
	selectedDateTo := ""

	if selectedQ != "" {
		filters["q"] = selectedQ
	}
	if t := c.Query("type"); t != "" && service.IsValidItemType(t) {
		filters["type"] = t
		selectedType = t
	}
	if st := c.Query("status"); st != "" && service.IsValidApprovalStatus(st) {
		filters["approval_status"] = st
		selectedStatus = st
	}
	if cgy := c.Query("category"); cgy != "" && service.IsValidCategory(cgy) {
		filters["category"] = cgy
		selectedCategory = cgy
	}
	if loc := c.Query("location"); loc != "" && service.IsValidASTULocation(loc) {
		filters["location"] = loc
		selectedLocation = loc
	}
	if from := c.Query("date_from"); from != "" {
		if _, err := time.Parse("2006-01-02", from); err == nil {
			filters["date_from"] = from
			selectedDateFrom = from
		}
	}
	if to := c.Query("date_to"); to != "" {
		if _, err := time.Parse("2006-01-02", to); err == nil {
			filters["date_to"] = to
			selectedDateTo = to
		}
	}

	items, _ := h.itemService.SearchItems(filters)
	typedUser, _ := user.(model.User)
	unreadCount := h.itemService.CountUnreadNotifications(typedUser.ID)

	renderHTML(c, http.StatusOK, "admin_items.html", gin.H{
		"title":              "Manage Item Posts",
		"user":               user,
		"items":              items,
		"filters":            filters,
		"locations":          service.ASTULocations(),
		"categories":         service.ItemCategories(),
		"selected_q":         selectedQ,
		"selected_type":      selectedType,
		"selected_status":    selectedStatus,
		"selected_category":  selectedCategory,
		"selected_location":  selectedLocation,
		"selected_date_from": selectedDateFrom,
		"selected_date_to":   selectedDateTo,
		"unread_count":       unreadCount,
		"content_template":   "admin_items_content",
	})
}

func (h *AdminHandler) UpdateClaim(c *gin.Context) {
	claimID, _ := strconv.ParseUint(c.PostForm("claim_id"), 10, 32)
	status := c.PostForm("status")
	remarks := c.PostForm("remarks")

	err := h.itemService.UpdateClaimStatus(uint(claimID), status, remarks)
	if err != nil {
		renderHTML(c, http.StatusOK, "error.html", gin.H{
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
		renderHTML(c, http.StatusOK, "error.html", gin.H{
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
		renderHTML(c, http.StatusOK, "error.html", gin.H{
			"error":            "Failed to remove item: " + err.Error(),
			"title":            "Error",
			"content_template": "error_content",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/items")
}
