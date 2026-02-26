// internal/handler/admin_handler.go
package handler

import (
    "net/http"
    "strconv"
    "lostfound/internal/middleware"
    "lostfound/internal/service"
    
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

// Show admin dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
    user, _ := c.Get("user")
    stats, _ := h.itemService.GetStats()
    claims, _ := h.itemService.GetAllClaims()
    
    c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
        "title": "Admin Dashboard",
        "user": user,
        "stats": stats,
        "claims": claims,
    })
}

// Show claims management
func (h *AdminHandler) ShowClaims(c *gin.Context) {
    claims, _ := h.itemService.GetAllClaims()
    
    c.HTML(http.StatusOK, "admin/claims.html", gin.H{
        "title": "Manage Claims",
        "claims": claims,
    })
}

// Handle claim approval/rejection
func (h *AdminHandler) UpdateClaim(c *gin.Context) {
    claimID, _ := strconv.ParseUint(c.PostForm("claim_id"), 10, 32)
    status := c.PostForm("status")
    remarks := c.PostForm("remarks")
    
    err := h.itemService.UpdateClaimStatus(uint(claimID), status, remarks)
    if err != nil {
        c.HTML(http.StatusOK, "error.html", gin.H{
            "error": "Failed to update claim: " + err.Error(),
        })
        return
    }
    
    c.Redirect(http.StatusSeeOther, "/admin/claims")
}