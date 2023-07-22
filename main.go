package main

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/jgsheppa/gin-playground/models"
	"github.com/jgsheppa/gin-playground/server"
	"github.com/spf13/viper"
)

func init() {
	configInit()
}

func main() {
	sentrySetup()

	dbName := viper.GetString("DATABASE_NAME")
	s := models.NewServices(dbName)
	err := s.AutoMigrate()
	if err != nil {
		log.Fatal("Could not migrate database: %w", err)
	}
	r := server.RunServer(s)
	// Run the server per default on port 8080
	r.Run()
}

func configInit() {
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

func sentrySetup() {
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
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
}
