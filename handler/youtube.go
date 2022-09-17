package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func UpdateYoutubeVideos(ctx *gin.Context) {
	yctx := context.Background()
	yts, err := youtube.NewService(yctx, option.WithAPIKey("AIzaSyD41vI8OY-zVgU1QIFYrW43xl7UbIz75uc"))
	if err != nil {
		fmt.Errorf("error ", err)
	}
	call := yts.Search.List([]string{"snippet"}).Q("cricket").Type("video").MaxResults(5).PublishedBefore(time.Now().Format(time.RFC3339)).Order("relevance")
	resp, err := call.Do()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err})
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}
