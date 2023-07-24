package server

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/controllers"
	"github.com/jgsheppa/gin-playground/middleware"
	"github.com/jgsheppa/gin-playground/models"
)

func RunServer(s *models.Services) *gin.Engine {
	fileController := controllers.NewFile(s.File)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.LoadHTMLGlob("templates/*")
	r.Use(sentrygin.New(sentrygin.Options{}))
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Recovery())

	r.GET("/", fileController.GetFiles)

	file := r.Group("/file")
	{
		file.POST("/upload", fileController.Upload)
		file.POST("/download", fileController.Download)
		file.POST("/delete", fileController.Delete)
	}

	return r
}
