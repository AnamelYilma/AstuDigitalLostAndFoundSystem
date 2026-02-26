package devkit

import (
	_ "github.com/gofiber/fiber/v2"
	_ "github.com/gin-gonic/gin"
	_ "github.com/labstack/echo/v4"
	_ "gorm.io/gorm"
	_ "gorm.io/driver/postgres"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv"
	_ "github.com/spf13/viper"
	_ "github.com/go-resty/resty/v2"
	_ "github.com/sirupsen/logrus"
	_ "go.uber.org/zap"
	_ "github.com/redis/go-redis/v9"
)
