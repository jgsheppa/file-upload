package main

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/controllers"
	"github.com/jgsheppa/gin-playground/models"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("yaml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/secrets") // Used for deployments to Render
	viper.AddConfigPath(".")            // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatalf("config.yaml not found %v", err)
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error loading config file: %w", err))
		}
	}
}

func main() {
	sentryInit()
	runServer()
}

func sentryInit() {
	sentryKey := viper.GetString("SENTRY_KEY")
	environment := viper.GetString("ENVIRONMENT")
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
}

func runServer() {
	dbName := viper.GetString("DATABASE_NAME")

	s := models.NewServices(dbName + ".db")
	err := s.AutoMigrate()
	if err != nil {
		log.Fatal("Could not migrate database: %w", err)
	}

	fileController := controllers.NewFile(s.File)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.LoadHTMLGlob("templates/*")
	r.Use(sentrygin.New(sentrygin.Options{}))

	r.GET("/", fileController.GetFiles)

	file := r.Group("/file")
	{
		file.POST("/upload", fileController.Upload)
		file.POST("/download", fileController.Download)
		file.POST("/delete", fileController.Delete)
	}

	// Run the server per default on port 8080
	r.Run(":8080")
}
