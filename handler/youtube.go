package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nitinskumavat/youtube-fetch/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var next_page_token string

func fetchFromYoutube(query, content_type, published_before, next_page_token string) (youtube.SearchListResponse, error) {
	ctx := context.Background()
	yts, err := youtube.NewService(ctx, option.WithAPIKey("AIzaSyD41vI8OY-zVgU1QIFYrW43xl7UbIz75uc"))
	if err != nil {
		return youtube.SearchListResponse{}, err
	}
	call := yts.Search.List([]string{"snippet"}).Q(query).Type(content_type).MaxResults(50).PublishedBefore(published_before).Order("date").PageToken(next_page_token)
	resp, err := call.Do()
	if err != nil {
		return youtube.SearchListResponse{}, err
	}
	return *resp, nil
}

func UpdateYoutubeVideos(ctx *gin.Context) {
	yctx := context.Background()
	yts, err := youtube.NewService(yctx, option.WithAPIKey("AIzaSyD41vI8OY-zVgU1QIFYrW43xl7UbIz75uc"))
	if err != nil {
		log.Fatal("error ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err})
	}
	// t:=time.Now().Format(time.RFC3339)
	call := yts.Search.List([]string{"snippet"}).Q("iphone").Type("video").MaxResults(50).PublishedAfter("2010-01-01T00:00:00Z").Order("date")
	resp, err := call.Do()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	models := make([]mongo.WriteModel, 0)
	for _, item := range resp.Items {
		t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		fmt.Println("timexxxxx ", t)
		fmt.Println("timeyyyyy", t.Format(time.RFC3339))
		model := mongo.NewUpdateOneModel()
		model.SetUpsert(true)
		model.SetFilter(bson.D{{Key: "youtube_id", Value: item.Id.VideoId}})
		item_update_data := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "youtube_id", Value: item.Id.VideoId},
				{Key: "title", Value: item.Snippet.Title},
				{Key: "description", Value: item.Snippet.Description},
				{Key: "thumbnail", Value: item.Snippet.Thumbnails.Default.Url},
				{Key: "published_at", Value: primitive.NewDateTimeFromTime(t)},
			}},
		}
		model.SetUpdate(item_update_data)
		fmt.Println(model)
		models = append(models, model)
	}
	res, _ := database.InsertManyItemToDB(models)
	ctx.JSON(http.StatusOK, gin.H{"data": res})
}

func PrimitiveDateToUtcString(date_time primitive.DateTime) string {
	return date_time.Time().UTC().Format(time.RFC3339)
}

func Recurse() {
	youtube_data, err := fetchFromYoutube("iphone", "video", time.Now().Format(time.RFC3339), next_page_token)
	fmt.Println("--youtube-data----", youtube_data.NextPageToken, " ", next_page_token)
	next_page_token = youtube_data.NextPageToken
	fmt.Println("youtube_data ", len(youtube_data.Items))
	if err != nil {
		log.Fatal(err)
	}
	models := make([]mongo.WriteModel, 0)
	for _, item := range youtube_data.Items {
		t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		model := mongo.NewUpdateOneModel()
		model.SetUpsert(true)
		model.SetFilter(bson.D{{Key: "youtube_id", Value: item.Id.VideoId}})
		model_data := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "youtube_id", Value: item.Id.VideoId},
				{Key: "title", Value: item.Snippet.Title},
				{Key: "description", Value: item.Snippet.Description},
				{Key: "thumbnail", Value: item.Snippet.Thumbnails.Default.Url},
				{Key: "published_at", Value: primitive.NewDateTimeFromTime(t)},
			}},
		}
		model.SetUpdate(model_data)
		models = append(models, model)
	}
	fmt.Println(len(models))
	fmt.Printf("models: %v\n", models)
	resp, err := database.InsertManyItemToDB(models)
	if err != nil {
		fmt.Println("Error Inserting data ", err)
	}
	fmt.Println("Inserted successful with inserted count ", resp)
}

func UpdateLatestVideos() {
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				Recurse()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
