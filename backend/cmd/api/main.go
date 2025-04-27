package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/handlers"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/middleware"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/service"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	listenAddr := flag.String("listen", "", "address to listen on")
	flag.Parse()

	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
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

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, viper.GetString("app.secret_key"))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	runnerHandler := handlers.NewRunnerHandler(submissionRepo, problemRepo)

	r := gin.Default()

	// Public routes
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// Authorized routes
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(viper.GetString("app.secret_key")))
	{
		//user
		admin := authorized.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			//admin
		}
	}

	// Internal API for runner service
	internalAPI := r.Group("/internal")
	internalAPI.Use(middleware.InternalAPIMiddleware(viper.GetString("runner.api_token")))
	{
		internalAPI.GET("/submissions/next", runnerHandler.GetNextSubmission)
		internalAPI.PUT("/submissions/:id/result", runnerHandler.UpdateSubmissionResult)
	}

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	} else if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
