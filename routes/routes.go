package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func SubmitDownload(c *gin.Context) {
	manager.SubmitDownload(c)
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

		downloads := manager.GetActiveDownloads()
		errors := manager.GetErrors()

		err = conn.WriteJSON(gin.H{
			"downloads": downloads,
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
