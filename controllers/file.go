package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/jgsheppa/gin-playground/models"
)

type File struct {
	fs models.FileService
}

func NewFile(fs models.FileService) *File {
	return &File{fs: fs}
}

func (f *File) Upload(c *gin.Context) {
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		sentry.CaptureException(errors.New("could read multipart form for UploadMultiple endpoint"))
		c.String(http.StatusBadRequest, fmt.Sprintf("could read multipart form for UploadMultiple endpoint: %s", err.Error()))
		return
	}
	files := form.File["upload[]"]

	for _, file := range files {
		content, err := file.Open()
		if err != nil {
			sentry.CaptureException(errors.New("could not open file for UploadMultiple endpoint"))
			c.String(http.StatusBadRequest, fmt.Sprintf("could not open file for UploadMultiple endpoint: %s", err.Error()))
			return
		}
		defer content.Close()

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, content); err != nil {
			sentry.CaptureException(errors.New("could not copy buffer for UploadMultiple endpoint"))
			c.String(http.StatusBadRequest, fmt.Sprintf("could not copy buffer for UploadMultiple endpoint: %s", err.Error()))
			return
		}

		err = f.fs.Create(&models.File{Filename: file.Filename, FileBlob: buf.Bytes()})
		if err != nil {
			sentry.CaptureException(fmt.Errorf("could not create file %s for UploadMultiple endpoint", file.Filename))
			c.String(http.StatusBadRequest, fmt.Sprintf("could not create file for UploadMultiple endpoint: %s", err.Error()))
			return
		}
	}
	c.Redirect(http.StatusFound, "/")
}

func (f *File) Download(c *gin.Context) {
	id := c.Query("id")
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("could not download file: %w", err))
		c.String(http.StatusBadRequest, fmt.Sprintf("could not download file %s:", err.Error()))
		return
	}
	file, err := f.fs.Find(idAsInt)

	c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
	http.ServeContent(c.Writer, c.Request, file.Filename, file.UpdatedAt, bytes.NewReader(file.FileBlob))
	c.Redirect(http.StatusFound, "/")
}

func (f *File) Delete(c *gin.Context) {
	id := c.Query("id")
	conv, err := strconv.Atoi(id)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("could not delete file: %w", err))
		c.String(http.StatusBadRequest, fmt.Sprintf("could not delete file %s:", err.Error()))
		return
	}

	err = f.fs.Delete(conv)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("could not delete file: %w", err))
		c.String(http.StatusBadRequest, fmt.Sprintf("could not delete file %s:", err.Error()))
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func (f *File) GetFiles(c *gin.Context) {
	files, err := f.fs.GetAll()
	if err != nil {
		sentry.CaptureException(fmt.Errorf("could not retrieve files: %w", err))
		c.String(http.StatusBadRequest, fmt.Sprintf("could not retrieve files %s:", err.Error()))
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":   "File Upload",
		"uploads": files,
	})
}
