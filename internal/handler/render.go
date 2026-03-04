package handler

import (
	"lostfound/internal/model"
	"lostfound/internal/service"
	"lostfound/pkg/database"

	"github.com/gin-gonic/gin"
)

func renderHTML(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	if token, ok := c.Get("csrf_token"); ok {
		if _, exists := data["csrf_token"]; !exists {
			data["csrf_token"] = token
		}
	}
	if userVal, ok := c.Get("user"); ok {
		if _, exists := data["user"]; !exists {
			data["user"] = userVal
		}
		if u, ok := userVal.(model.User); ok {
			if _, exists := data["unread_count"]; !exists {
				var unread int64
				database.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", u.ID, false).Count(&unread)
				data["unread_count"] = unread
			}
		}
	}
	// Provide common option lists if a template forgets to set them.
	if _, exists := data["locations"]; !exists {
		data["locations"] = service.ASTULocations()
	}
	if _, exists := data["categories"]; !exists {
		data["categories"] = service.ItemCategories()
	}
	if _, exists := data["colors"]; !exists {
		data["colors"] = service.ColorOptions()
	}
	c.HTML(status, name, data)
}
