package handler

import (
	"lostfound/internal/model"
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
	c.HTML(status, name, data)
}
