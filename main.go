package main

import (
	"fmt"
	"megazen/routes"

	"github.com/gin-gonic/gin"
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
	fmt.Println("MegaZen can now be accessed at http://localhost:3000")
}
