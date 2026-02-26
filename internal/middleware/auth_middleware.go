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
