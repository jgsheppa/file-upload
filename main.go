package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/models"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
func main() {
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
	err := s.AutoMigrate()
	if err != nil {
		log.Fatal("Could not migrate database: %w", err)
	}

	r := gin.Default()

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
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// HTTP redirect example
	r.GET("/httpRedirect", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/httpRedirect2")
	})
	r.GET("/httpRedirect2", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "world"})
	})

	// Router redirect example
	r.GET("/routerRedirect", func(c *gin.Context) {
		c.Request.URL.Path = "/routerRedirect2"
		r.HandleContext(c)
	})
	r.GET("/routerRedirect2", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "world"})
	})

	// Upload file example. Use curl to test it.
	//
	// curl -X POST http://localhost:8080/upload \
	// -F "file=@/Users/appleboy/test.zip" \
	// -H "Content-Type: multipart/form-data"
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.POST("/upload", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("upload")
		if err != nil {
			log.Println(err)
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		content, err := file.Open()
		defer content.Close()

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, content); err != nil {
			log.Fatal(err)
		}

		err = s.File.CreateFile(&models.File{Filename: file.Filename, FileBlob: buf.Bytes()})

		// Upload the file to specific dst. TODO: use with BLOB storage.
		// c.SaveUploadedFile(file, "./uploads/"+file.Filename)
		// FIXME: The redirect only seems to work with this status code.
		c.Redirect(http.StatusMovedPermanently, "/")
	})

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

	r.GET("/download", func(c *gin.Context) {
		id := c.Query("id")
		idAsInt, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}
		file, err := s.File.Get(idAsInt)

		fmt.Printf("file: %v", file)
	})

	// Multiple file upload example. Use curl to test it.
	// Similar to single file upload but with a loop.
	//
	// curl -X POST http://localhost:8080/upload \
	// -F "upload[]=@/Users/appleboy/test1.zip" \
	// -F "upload[]=@/Users/appleboy/test2.zip" \
	// -H "Content-Type: multipart/form-data"
	r.POST("/uploadMultiple", func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload[]"]

		for _, file := range files {
			log.Println(file.Filename)

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, ".")
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})

	// Capture a Sentry error when there is a 500 error.
	r.GET("/error", func(c *gin.Context) {
		sentry.CaptureMessage("An Error!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})

	// Sentry recovers from panic errors
	// and creates an issue, if a issues for
	// fatal errors have been set.
	r.GET("/fatal", func(c *gin.Context) {
		defer sentry.Recover()
		panic("A Fatal Error!")
	})

	// Cookie example
	r.GET("/cookie", func(c *gin.Context) {

		cookie, err := c.Cookie("gin_cookie")

		if err != nil {
			cookie = "NotSet"
			c.SetCookie("gin_cookie", "test", 3600, "/", "localhost", false, true)
		}

		fmt.Printf("Cookie value: %s \n", cookie)
	})

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe.
	// Check out http://localhost:8080/welcome?firstname=Jane&lastname=Doe
	r.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})

	// Grouping routes example
	// Simple group: v1
	v1 := r.Group("/v1")
	{
		v1.GET("/login", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "login1",
			})
		})
		v1.GET("/submit", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "submit1",
			})
		})
		v1.GET("/read", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "read1",
			})
		})
	}

	// Simple group: v2
	v2 := r.Group("/v2")
	{
		v2.GET("/login", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "login2",
			})
		})
		v2.GET("/submit", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "submit2",
			})
		})
		v2.GET("/read", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "read2",
			})
		})
	}

	// Run the server per default on port 8080
	r.Run()
}
