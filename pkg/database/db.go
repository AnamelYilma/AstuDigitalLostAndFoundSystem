package database
import (
	"gorm.io/gorm"
	"github.com/joho/godotenv"
	"log"


)


var DB *gorm.DB
func Intdb(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Warring : env file is not found or open")
	}
	dns = 


}