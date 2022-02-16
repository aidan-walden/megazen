package routes

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"megazen/models"
	"megazen/models/download-controller"
	"net/http"
)

var manager = download_controller.New(4)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func SubmitDownloadBulk(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	reqBody, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error: " + err.Error(),
		})
		return
	}

	urls := make([]string, 0)

	err = json.Unmarshal(reqBody, &urls)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad Request",
		})
		return
	}

	submissions := make([]models.DownloadSubmission, 0)

	for _, url := range urls {
		sub := models.DownloadSubmission{
			Url:      url,
			Password: "",
		}

		submissions = append(submissions, sub)
	}

	unknownUrls := manager.SubmitDownload(&submissions)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Downloads submitted",
		"unknownUrls": unknownUrls,
	})
}

func SubmitDownload(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	reqBody, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error: " + err.Error(),
		})
		return
	}

	urls := make([]models.DownloadSubmission, 0)

	err = json.Unmarshal(reqBody, &urls)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad Request",
		})
		return
	}

	unknownUrls := manager.SubmitDownload(&urls)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Downloads submitted",
		"unknownUrls": unknownUrls,
	})
}

func DownloadsWebsocket(c *gin.Context) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err := c.AbortWithError(http.StatusInternalServerError, err)
		if err != nil {
			panic(err)
		}
	}

	for {

		_, _, err := conn.ReadMessage()
		if err != nil {
			err := c.AbortWithError(http.StatusInternalServerError, err)
			if err != nil {
				panic(err)
			}
		}

		errors := make([]string, 0)

		downloads := manager.GetActiveDownloads()
		waiting := manager.GetWaitingDownloads()
		for _, err := range manager.GetErrors() {
			errors = append(errors, (*err).Error())
		}

		err = conn.WriteJSON(gin.H{
			"downloads": downloads,
			"waiting":   waiting,
			"errors":    errors,
		})

		if err != nil {
			err := c.AbortWithError(http.StatusInternalServerError, err)
			if err != nil {
				panic(err)
			}
		}
	}
}
