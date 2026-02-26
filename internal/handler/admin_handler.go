package handler

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

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"title":            "Admin Dashboard",
		"user":             user,
		"stats":            stats,
		"claims":           claims,
		"content_template": "admin_dashboard_content",
	})
}

func (h *AdminHandler) ShowClaims(c *gin.Context) {
	claims, _ := h.itemService.GetAllClaims()

	c.HTML(http.StatusOK, "admin/claims.html", gin.H{
		"title":            "Manage Claims",
		"claims":           claims,
		"content_template": "admin_claims_content",
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
