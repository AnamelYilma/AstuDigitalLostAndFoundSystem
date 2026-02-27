package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfSessionKey = "csrf_token"

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		token, _ := session.Values[csrfSessionKey].(string)
		if token == "" {
			token = randomToken(32)
			session.Values[csrfSessionKey] = token
			_ = session.Save(c.Request, c.Writer)
		}

		c.Set("csrf_token", token)

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			sent := c.PostForm("csrf_token")
			if sent == "" {
				sent = c.GetHeader("X-CSRF-Token")
			}
			if sent == "" || subtle.ConstantTimeCompare([]byte(sent), []byte(token)) != 1 {
				c.HTML(http.StatusForbidden, "error.html", gin.H{
					"title":            "Security Error",
					"error":            "Invalid CSRF token. Refresh and try again.",
					"content_template": "error_content",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
