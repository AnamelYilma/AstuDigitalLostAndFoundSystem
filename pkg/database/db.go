package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	var dsn string

	if dbURL != "" {
		dsn = dbURL
	} else {
		host := getEnv("DB_HOST", "localhost")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "")
		name := getEnv("DB_NAME", "lostfound")
		port := getEnv("DB_PORT", "5432")
		sslmode := getEnv("DB_SSLMODE", "disable")

		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, name, port, sslmode,
		)
	}

	logDSN := hidePassword(dsn)
	log.Printf("Connecting with: %s", logDSN)

	var err error
	logLevel := logger.Info
	if strings.EqualFold(os.Getenv("GO_ENV"), "production") {
		logLevel = logger.Warn
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")
}

func fixRenderURL(url string) string {
	if strings.HasPrefix(url, "postgresql://") {
		url = strings.Replace(url, "postgresql://", "postgres://", 1)
	}

	if strings.Contains(url, "postgres://") && !strings.Contains(url, ":5432") {
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			hostPart := parts[1]
			hostPortDb := strings.SplitN(hostPart, "/", 2)
			if len(hostPortDb) == 2 {
				hostPort := hostPortDb[0]
				dbName := hostPortDb[1]
				if !strings.Contains(hostPort, ":") {
					url = parts[0] + "@" + hostPort + ":5432/" + dbName
				}
			}
		}
	}

	if !strings.Contains(url, "sslmode=") {
		if strings.Contains(url, "?") {
			url += "&sslmode=require"
		} else {
			url += "?sslmode=require"
		}
	}

	return url
}

func hidePassword(dsn string) string {
	if strings.Contains(dsn, "postgres://") {
		return "postgres://****:****@" + strings.SplitN(dsn, "@", 2)[1]
	}
	if strings.Contains(strings.ToLower(dsn), "password=") {
		parts := strings.Fields(dsn)
		for i, p := range parts {
			if strings.HasPrefix(strings.ToLower(p), "password=") {
				parts[i] = "password=****"
			}
		}
		return strings.Join(parts, " ")
	}
	return dsn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
