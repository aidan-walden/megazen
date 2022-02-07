package routes

import (
	"github.com/gin-gonic/gin"
	"megazen/models/download-controller"
	"net/http"
)

func Index(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

var manager = download_controller.New()

func SubmitDownload(c *gin.Context) {
	manager.SubmitDownload(c)
}
