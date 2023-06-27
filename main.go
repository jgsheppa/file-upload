package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// HTML rendering example
	r.LoadHTMLGlob("templates/*")
	//r.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	r.GET("/", func(c *gin.Context) {
		os.Chmod("./uploads/", 0777)
		files, err := os.ReadDir("./uploads")
		if err != nil {
			log.Println(err)
			c.String(http.StatusBadRequest, fmt.Sprintf("got os err: %s", err.Error()))
			return
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
		log.Println(file.Filename)

		// Upload the file to specific dst.
		c.SaveUploadedFile(file, "./uploads/"+file.Filename)
		// FIXME: The redirect only seems to work with this status code.
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// Deleting files from a server.
	r.POST("/deleteUploads", func(c *gin.Context) {
		fileName := c.Query("fileName")
		files, err := os.ReadDir("./uploads")
		if err != nil {
			log.Println(err)
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		for _, file := range files {
			if file.Name() == fileName {
				err := os.Remove("./uploads/" + fileName)
				if err != nil {
					log.Println(err)
					c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
					return
				}
				// FIXME: The redirect only seems to work with this status code.
				c.Redirect(http.StatusMovedPermanently, "/")
				return
			}
		}

		c.String(http.StatusNotFound, fmt.Sprintf("file not found"))
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
