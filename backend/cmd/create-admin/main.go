package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	username := flag.String("username", "", "admin username")
	password := flag.String("password", "", "admin password")
	email := flag.String("email", "", "admin email")
	flag.Parse()

	// Validate required flags
	if *username == "" || *password == "" || *email == "" {
		log.Fatal("Username, password, and email are required")
	}

	// Read config
	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetString("database.sslmode"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	// Check if user exists
	var existingUser models.User
	result := db.Where("username = ?", *username).First(&existingUser)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Fatalf("Error checking for existing user: %s", result.Error)
	}

	if result.RowsAffected == 0 {
		// User doesn't exist, create a new admin user
		user := models.User{
			Username: *username,
			Password: *password, // Will be hashed by BeforeSave hook
			Email:    *email,
			Role:     models.RoleAdmin,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create admin user: %s", err)
		}

		log.Printf("Admin user '%s' created successfully", *username)
	} else {
		// User exists, update to admin role if not already
		if existingUser.Role != models.RoleAdmin {
			existingUser.Role = models.RoleAdmin
			if err := db.Save(&existingUser).Error; err != nil {
				log.Fatalf("Failed to update user role: %s", err)
			}
			log.Printf("User '%s' role updated to admin", *username)
		} else {
			log.Printf("User '%s' is already an admin", *username)
		}
	}
}
