package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/runner"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	tempDir := flag.String("temp-dir", "", "directory for temporary files")
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Read config
	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatalf("Error reading config file: %s", err)
	}

	// Get API details from config
	apiURL := viper.GetString("runner.api_url")
	if apiURL == "" {
		apiURL = "http://localhost:8080/internal"
	}
	apiToken := viper.GetString("runner.api_token")
	if apiToken == "" {
		logger.Fatal("API token is required in config")
	}

	// Create runner
	r, err := runner.NewRunner(*tempDir)
	if err != nil {
		logger.Fatalf("Failed to create runner: %s", err)
	}

	// Create API client for backend communication
	client := runner.NewAPIClient(apiURL, apiToken, logger)

	// Create processor
	processor := runner.NewProcessor(client, r, logger)

	// Run the processor
	logger.Info("Starting runner service")
	stopCh := make(chan struct{})
	go processor.Start(stopCh)

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down runner service")
	close(stopCh)

	// Allow some time for graceful shutdown
	time.Sleep(2 * time.Second)
}
