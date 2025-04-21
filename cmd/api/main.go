package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/yousefi-abolfazl/code-judge/internal/middleware"
	"github.com/yousefi-abolfazl/code-judge/internal/models"
	"github.com/yousefi-abolfazl/code-judge/internal/repository"
	"github.com/yousefi-abolfazl/code-judge/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	listenAddr := flag.String("listen", "", "address to listen on")
	flag.Parse()

	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Feralf("Failed to run migrations: %s", err)
	}

	if *listenAddr != "" {
		viper.Set("app.port", *listenAddr)
	}

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

	err = db.AutoMigrate(&models.User{}, &models.Problem{}, &models.Submission{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepo, viper.GetString("app.secret_key"))

	authHandler := handlers.NewAuthHandler(authService)

	r := gin.Default()

	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(viper.GetString("app.secret_key")))
	{
		// user
		admin := authorized.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			// admin
		}
	}

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	} else if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Feralf("Failed to start server: %s", err)
	}
}
