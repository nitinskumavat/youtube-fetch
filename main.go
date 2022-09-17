package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nitinskumavat/youtube-fetch/handler"
)

func main() {
	r := gin.Default()
	r.GET("/", handler.UpdateYoutubeVideos)
	r.Run(":3001")
}
