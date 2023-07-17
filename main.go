package main

import (
	"fmt"
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/controllers"
	"github.com/jgsheppa/gin-playground/models"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}
	err = godotenv.Load("/etc/secrets/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}
	sentryKey := os.Getenv("SENTRY_KEY")
	environment := os.Getenv("ENVIRONMENT")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           sentryKey,
		Environment:   environment,
		Release:       "playground@1.0.0", // TODO: use git commit hash
		EnableTracing: true,
		Debug:         true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	s := models.NewServices("dev.db")
	err = s.AutoMigrate()
	if err != nil {
		log.Fatal("Could not migrate database: %w", err)
	}
	runServer(s)
}

func runServer(s *models.Services) {
	fileController := controllers.NewFile(s.File)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.LoadHTMLGlob("templates/*")
	r.Use(sentrygin.New(sentrygin.Options{}))

	// Could be used in an endpoint to identify which users
	// have had issues with specific endpoints.
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: "jane.doe@example.com"})
	})

	r.GET("/", fileController.GetFiles)

	file := r.Group("/file")
	{
		file.POST("/upload", fileController.Upload)
		file.POST("/download", fileController.Download)
		file.POST("/delete", fileController.Delete)
	}

	// Run the server per default on port 8080
	r.Run(":8081")
}
