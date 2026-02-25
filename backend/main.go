package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// -- Models --

type User struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type Item struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	ReporterID  uint      `json:"reporter_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// -- Database (In-Memory) --

var (
	users      = make(map[string]*User)
	items      = make(map[uint]*Item)
	itemCounter uint = 1
	mu         sync.Mutex
	jwtSecret  = []byte("your-very-secret-key")
)

func init() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	users["admin@test.com"] = &User{ID: 1, Email: "admin@test.com", Password: string(hash), Role: "admin"}

	hash2, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	users["user@test.com"] = &User{ID: 2, Email: "user@test.com", Password: string(hash2), Role: "user"}
}

// -- Auth Logic --

func generateToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// -- Middleware --

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("userId", uint(claims["sub"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

func RBACMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != requiredRole && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// -- Handlers --

func loginHandler(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	mu.Lock()
	user, ok := users[input.Email]
	mu.Unlock()

	if !ok || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := generateToken(user)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", token, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, user)
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func meHandler(c *gin.Context) {
	userId, _ := c.Get("userId")
	mu.Lock()
	defer mu.Unlock()
	for _, u := range users {
		if u.ID == userId {
			c.JSON(http.StatusOK, u)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func getItems(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	var itemList []Item
	for _, item := range items {
		itemList = append(itemList, *item)
	}
	c.JSON(http.StatusOK, itemList)
}

func createItem(c *gin.Context) {
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	userId, _ := c.Get("userId")
	item.ReporterID = userId.(uint)
	item.CreatedAt = time.Now()
	mu.Lock()
	item.ID = itemCounter
	items[itemCounter] = &item
	itemCounter++
	mu.Unlock()
	c.JSON(http.StatusCreated, item)
}

func deleteItem(c *gin.Context) {
	var input struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	mu.Lock()
	delete(items, input.ID)
	mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

func main() {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // Vite default
	config.AllowCredentials = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	r.Use(cors.New(config))

	auth := r.Group("/auth")
	{
		auth.POST("/login", loginHandler)
		auth.POST("/logout", logoutHandler)
		auth.GET("/me", AuthMiddleware(), meHandler)
	}

	api := r.Group("/api")
	api.Use(AuthMiddleware())
	{
		api.GET("/items", getItems)
		api.POST("/items", createItem)
		api.DELETE("/items/:id", RBACMiddleware("admin"), deleteItem)
	}

	fmt.Println("Server starting on :8080")
	r.Run(":8080")
}
