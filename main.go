package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nitinskumavat/youtube-fetch/database"
)

func main() {
	fmt.Println("Starting the application....")
	collection := database.ConnectToDB()
	fmt.Println(collection)
	r := gin.Default()
	r.GET("/search", database.GetQueryVideos)
	r.GET("/videos", database.GetVideos)
	r.GET("/deletemany", database.DeleteMany)
	// go handler.UpdateLatestVideos()
	r.Run(":3000")
}
