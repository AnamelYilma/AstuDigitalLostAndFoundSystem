package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"lostfound/internal/model"
	"lostfound/pkg/database"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var (
	store     *sessions.CookieStore
	storeOnce sync.Once
)

func initStore() {
	store = sessions.NewCookieStore([]byte(getSessionSecret()))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		SameSite: http.SameSiteLaxMode,
	}
}

func getSessionSecret() string {
	if v := os.Getenv("SESSION_SECRET"); v != "" {
		if len(v) >= 32 {
			return v
		}
		sum := sha256.Sum256([]byte(v))
		log.Println("SESSION_SECRET is shorter than 32 chars; using SHA-256 derived key")
		return base64.StdEncoding.EncodeToString(sum[:])
	}
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err == nil {
		log.Println("SESSION_SECRET not set; generated an ephemeral secret for this run")
		return base64.StdEncoding.EncodeToString(buf)
	}
	log.Println("SESSION_SECRET not set and secure random failed; using fallback development secret")
	return "dev-only-session-secret-change-this-32chars"
}

func GetSession(c *gin.Context) *sessions.Session {
	storeOnce.Do(initStore)
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
