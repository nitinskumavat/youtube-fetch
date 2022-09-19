package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nitinskumavat/youtube-fetch/database"
	yt "github.com/nitinskumavat/youtube-fetch/youtube"
)

func main() {
	fmt.Println("Starting the application....")
	collection := database.ConnectToDB()
	fmt.Println(collection)
	r := gin.Default()
	r.GET("/search", database.GetQueryVideos)
	r.GET("/videos", database.GetVideos)
	// r.DELETE("/delete-all", database.DeleteMany)
	go yt.UpdateLatestVideos()
	r.Run(":12345")
}
