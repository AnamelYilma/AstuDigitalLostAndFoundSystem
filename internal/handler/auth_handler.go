package handler

import (
	"lostfound/internal/middleware"
	"lostfound/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title":            "Login",
		"content_template": "login_content",
	})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title":            "Register",
		"content_template": "register_content",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	studentID := c.PostForm("student_id")
	password := c.PostForm("password")

	user, err := h.authService.Login(studentID, password)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":            "Login",
			"error":            err.Error(),
			"content_template": "login_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	if user.Role == "admin" {
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	} else {
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	name := c.PostForm("name")
	studentID := c.PostForm("student_id")
	phone := c.PostForm("phone")
	password := c.PostForm("password")

	user, err := h.authService.Register(name, studentID, phone, password)
	if err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title":            "Register",
			"error":            err.Error(),
			"content_template": "register_content",
		})
		return
	}

	session := middleware.GetSession(c)
	session.Values["user_id"] = user.ID
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := middleware.GetSession(c)
	session.Values = make(map[interface{}]interface{})
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusSeeOther, "/")
}
