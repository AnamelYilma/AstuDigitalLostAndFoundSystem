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


	// Log (safely hide password)
	logDSN := hidePassword(dsn)
	log.Printf("Connecting with: %s", logDSN)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")
}

// FIX RENDER URL - CORRECTED VERSION
func fixRenderURL(url string) string {
	// 1. Replace postgresql:// with postgres:// for GORM compatibility
	if strings.HasPrefix(url, "postgresql://") {
		url = strings.Replace(url, "postgresql://", "postgres://", 1)
	}

	// 2. Ensure port 5432 is included if not present
	if strings.Contains(url, "postgres://") && !strings.Contains(url, ":5432") {
		// Check if port is missing after @
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			hostPart := parts[1]
			// Split by / to get host:port
			hostPortDb := strings.SplitN(hostPart, "/", 2)
			if len(hostPortDb) == 2 {
				hostPort := hostPortDb[0]
				dbName := hostPortDb[1]
				if !strings.Contains(hostPort, ":") {
					// Add port 5432
					url = parts[0] + "@" + hostPort + ":5432/" + dbName
				}
			}
		}
	}

	// 3. Add sslmode=require if not present (Render PostgreSQL requires SSL)
	if !strings.Contains(url, "sslmode=") {
		if strings.Contains(url, "?") {
			url += "&sslmode=require"
		} else {
			url += "?sslmode=require"
		}
	}

	return url
}

// HIDE PASSWORD IN LOGS - SIMPLIFIED
func hidePassword(dsn string) string {
	if strings.Contains(dsn, "postgres://") {
		return "postgres://****:****@" + strings.SplitN(dsn, "@", 2)[1]
	}
	return dsn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
