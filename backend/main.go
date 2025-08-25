// in backend/main.go

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/handlers"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/middleware"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/runner"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go [serve|runner|create-admin]")
	}

	command := os.Args[1]

	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	flag.Parse()

	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	switch command {
	case "serve":
		runServer()
	case "runner":
		runRunner()
	case "create-admin":
		createAdmin()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

// --- Server Logic ---
func runServer() {
	db := connectDB()
	migrateDB(db)

	userRepo := repository.NewUserRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	authService := service.NewAuthService(userRepo, viper.GetString("app.secret_key"))
	problemService := service.NewProblemService(problemRepo)
	submissionService := service.NewSubmissionService(submissionRepo, problemRepo)
	userService := service.NewUserService(userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	problemHandler := handlers.NewProblemHandler(problemService)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)
	userHandler := handlers.NewUserHandler(userService)
	runnerHandler := handlers.NewRunnerHandler(submissionRepo, problemRepo)

	r := gin.Default()

	// ... (تمام روت‌هایتان را اینجا کپی کنید)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(viper.GetString("app.secret_key")))
	{
		authorized.POST("/problems", problemHandler.CreateProblem)
		authorized.GET("/problems", problemHandler.GetPublishedProblems)
		authorized.GET("/problems/:id", problemHandler.GetProblemByID)
		authorized.POST("/submissions", submissionHandler.CreateSubmission)
		authorized.GET("/submissions/me", submissionHandler.GetMySubmissions)
		authorized.GET("/users/:id", userHandler.GetUserProfile)

		admin := authorized.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.PATCH("/users/:id/role", userHandler.UpdateUserRole)
			admin.GET("/problems", problemHandler.GetAllProblems)
			admin.PATCH("/problems/:id/status", problemHandler.UpdateProblemStatus)
		}
	}
	internalAPI := r.Group("/internal")
	internalAPI.Use(middleware.InternalAPIMiddleware(viper.GetString("runner.api_token")))
	{
		internalAPI.GET("/submissions/next", runnerHandler.GetNextSubmission)
		internalAPI.PUT("/submissions/:id/result", runnerHandler.UpdateSubmissionResult)
	}

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

// --- Runner Logic ---
func runRunner() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	apiURL := viper.GetString("runner.api_url")
	apiToken := viper.GetString("runner.api_token")
	dockerImage := viper.GetString("runner.docker_image")
	if apiToken == "" || dockerImage == "" || apiURL == "" {
		logger.Fatal("runner.api_url, runner.api_token, and runner.docker_image are required in config")
	}

	r, err := runner.NewRunner("/tmp/runner", dockerImage)
	if err != nil {
		logger.Fatalf("Failed to create runner: %s", err)
	}

	client := runner.NewAPIClient(apiURL, apiToken, logger)
	processor := runner.NewProcessor(client, r, logger)

	logger.Info("Starting runner service")
	stopCh := make(chan struct{})
	go processor.Start(stopCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down runner service")
	close(stopCh)
}

// --- Create Admin Logic ---
func createAdmin() {
	// این بخش از os.Args می‌خواند چون ساده‌تر است
	if len(os.Args) != 5 {
		log.Fatal("Usage: ... create-admin <username> <password> <email>")
	}
	username := os.Args[2]
	password := os.Args[3]
	email := os.Args[4]

	db := connectDB()

	var existingUser models.User
	result := db.Where("username = ?", username).First(&existingUser)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Fatalf("Error checking for existing user: %s", result.Error)
	}

	if result.RowsAffected == 0 {
		user := models.User{
			Username: username,
			Password: password, // Hashed by BeforeSave hook
			Email:    email,
			Role:     models.RoleAdmin,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create admin user: %s", err)
		}
		log.Printf("Admin user '%s' created successfully", username)
	} else {
		existingUser.Role = models.RoleAdmin
		if err := db.Save(&existingUser).Error; err != nil {
			log.Fatalf("Failed to update user role: %s", err)
		}
		log.Printf("User '%s' role updated to admin", username)
	}
}

// --- Helper Functions ---
func connectDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetString("database.sslmode"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	return db
}

func migrateDB(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{}, &models.Problem{}, &models.Submission{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}
}
