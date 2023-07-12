package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Redirect(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/httpRedirect2")
}

func RedirectDestination(c *gin.Context) {
	c.JSON(200, gin.H{"hello": "world"})
}

func RouterRedirectDestination(c *gin.Context) {
	c.JSON(200, gin.H{"hello": "world"})
}

func SentryError(c *gin.Context) {
	sentry.CaptureException(errors.New("An Error!"))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func SentryRecover(c *gin.Context) {
	defer sentry.Recover()
	panic("A Fatal Error!")
}

func Cookie(c *gin.Context) {

	cookie, err := c.Cookie("gin_cookie")

	if err != nil {
		cookie = "NotSet"
		c.SetCookie("gin_cookie", "test", 3600, "/", "localhost", false, true)
	}
	c.String(http.StatusOK, fmt.Sprintf("Here's your cookie's name: %s", cookie))
}

func QueryParams(c *gin.Context) {
	firstname := c.DefaultQuery("firstname", "Guest")
	lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

	c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
}

// Grouping routes example
// Simple group: v1
func V1(r *gin.Engine) {
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
}

// Simple group: v2
func V2(r *gin.Engine) {
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
}
