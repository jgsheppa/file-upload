package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Errors) == 0 {
			c.Next()
		}

		for _, ginErr := range c.Errors {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
				"title": ginErr.Err,
			})
		}
	}
}
