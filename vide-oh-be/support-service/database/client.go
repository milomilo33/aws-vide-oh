package database

import (
	"log"
	"support-service/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Instance *gorm.DB
var dbError error

func Connect(connectionString string) {
	Instance, dbError = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if dbError != nil {
		log.Fatal(dbError)
		panic("Cannot connect to DB")
	}
	log.Println("Connected to Database!")
}

func Migrate() {
	if !Instance.Migrator().HasTable(&models.Message{}) {
		Instance.AutoMigrate(&models.Message{})
		log.Println("Table 'messages' created and database migration completed!")
	} else {
		log.Println("Table 'messages' already exists, skipping migration.")
	}
}
