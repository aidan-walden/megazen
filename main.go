package main

import (
	"github.com/gin-gonic/gin"
	"megazen/routes"
)

func main() {
	r := gin.Default()

	r.LoadHTMLFiles("http/static/html/index.html")

	r.GET("/", routes.Index)

	api := r.Group("/api")
	{
		api.POST("/submit", routes.SubmitDownload)
		api.POST("/submitBulk", routes.SubmitDownloadBulk)
	}

	websocket := r.Group("/ws")
	{
		websocket.GET("/downloads", routes.DownloadsWebsocket)
	}

	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}
