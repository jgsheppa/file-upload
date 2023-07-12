package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/controllers"
	"github.com/jgsheppa/gin-playground/models"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
}
func main() {

	s := models.NewServices("dev.db")
	err := s.AutoMigrate()
	if err != nil {
		log.Fatal("Could not migrate database: %w", err)
	}

	fileController := controllers.NewFile(s.File)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.Use(sentrygin.New(sentrygin.Options{}))

	// Could be used in an endpoint to identify which users
	// have had issues with specific endpoints.
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: "jane.doe@example.com"})
	})

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	sentry.CaptureMessage("It works!")
	// HTML rendering example
	r.LoadHTMLGlob("templates/*")
	//r.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	r.GET("/", func(c *gin.Context) {
		files, err := s.File.GetAll()
		if err != nil {
			log.Fatal("Could not retrieve files: %w", err)
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":   "Main website",
			"uploads": files,
		})
	})

	// Basic ping example
	r.GET("/ping", controllers.Ping)

	// HTTP redirect example
	r.GET("/httpRedirect", controllers.Redirect)
	r.GET("/httpRedirect2", controllers.RedirectDestination)

	// Router redirect example
	r.GET("/routerRedirect", func(c *gin.Context) {
		c.Request.URL.Path = "/routerRedirect2"
		r.HandleContext(c)
	})
	r.GET("/routerRedirect2", controllers.RouterRedirectDestination)

	// YAML response example
	r.GET("/someYAML", func(c *gin.Context) {
		c.YAML(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})

	r.POST("/upload", fileController.Upload)
	r.POST("/download", fileController.Download)

	// Deleting files from a server.
	r.POST("/deleteUploads", func(c *gin.Context) {
		id := c.Query("id")
		conv, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}

		err = s.File.Delete(conv)
		if err != nil {
			log.Fatal(err)
		}

		// FIXME: The redirect only seems to work with this status code.
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// Capture a Sentry error when there is a 500 error.
	r.GET("/error", controllers.SentryError)

	// Sentry recovers from panic errors
	// and creates an issue, if a issues for
	// fatal errors have been set.
	r.GET("/fatal", controllers.SentryRecover)

	// Cookie example
	r.GET("/cookie", controllers.Cookie)

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe.
	// Check out http://localhost:8080/welcome?firstname=Jane&lastname=Doe
	r.GET("/welcome", controllers.QueryParams)

	controllers.V1(r)
	controllers.V2(r)

	// Run the server per default on port 8080
	r.Run(":3004")
}
