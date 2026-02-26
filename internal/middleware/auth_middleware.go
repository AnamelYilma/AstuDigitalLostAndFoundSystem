// internal/middleware/auth_middleware.go
package middleware

import (
    "net/http"
    "lostfound/pkg/database"
    "lostfound/internal/model"
    
    "github.com/gin-gonic/gin"
    "github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
    store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   86400 * 7, // 7 days
        HttpOnly: true,
        Secure:   false, // Set to true in production with HTTPS
    }
}

// Get session
func GetSession(c *gin.Context) *sessions.Session {
    session, _ := store.Get(c.Request, "auth-session")
    return session
}

// AuthRequired middleware
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := GetSession(c)
        userID, ok := session.Values["user_id"]
        
        if !ok {
            c.Redirect(http.StatusSeeOther, "/login")
            c.Abort()
            return
        }
        
        // Get user from database
        var user model.User
        if err := database.DB.First(&user, userID).Error; err != nil {
            session.Values = make(map[interface{}]interface{})
            session.Save(c.Request, c.Writer)
            c.Redirect(http.StatusSeeOther, "/login")
            c.Abort()
            return
        }
        
        // Set user in context
        c.Set("user", user)
        c.Next()
    }
}

// AdminRequired middleware
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
                "error": "Access denied. Admin only.",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// SetUser middleware (sets user if logged in)
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