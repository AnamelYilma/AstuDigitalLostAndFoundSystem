package database

import (
    "fmt"
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
    dsn := "host=localhost user=postgres password=0909 dbname=lostfound port=5432 sslmode=disable"
    
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        log.Fatal("❌ Failed to connect to database:", err)
    }
    
    fmt.Println("✅ Database connected successfully")
}