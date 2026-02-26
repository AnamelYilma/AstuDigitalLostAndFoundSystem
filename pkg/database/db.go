package database

import (
	"fmt"
	"log"
	"os"
	
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)



var DB *gorm.DB
func Intdb(){
	// err := godotenv.Load()
	dsn := "hostlocalhost user=postgres password=0909 dbname=lostfound port=5432 sslmode=disable"
	Db , err :=gorm.Open(postgres.Open(dsn), &gorm.Config{
					Logger: logger.Default.LogMode(logger.Info),
				})
	if err != nil {
		log.Fatalf("❌ warrning : Cannot connect to database!\n   Reason: %v\n   Check: Is PostgreSQL running? Is password correct? Does database 'app_go' exist?", err)
	}
	// Db,err := gorm.Open(postgres.Open(dsn) , &gorm.Config{})
	DB = Db
	fmt.Println("✅ Database connected successfully")

}