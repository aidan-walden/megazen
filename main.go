package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"megazen/routes"
)

func main() {
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./views", true)))

	api := r.Group("/api")
	{
		api.GET("/", routes.Index)
		api.POST("/submit", routes.SubmitDownload)
	}

	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}
