package handler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nitinskumavat/youtube-fetch/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var next_page_token string
var publish_time = time.Now().Format(time.RFC3339)
var api_key_index = 0
var cur_api_key string

func updateApiKeyAndIndex() {
	api_keys := strings.Split(os.Getenv("API_KEYS"), "#")
	cur_api_key = api_keys[api_key_index]
	api_key_index = (api_key_index + 1) % len(api_keys)
}

func fetchFromYoutube(query, content_type, published_before, next_page_token string) (youtube.SearchListResponse, error) {
	ctx := context.Background()
	fmt.Println("curr-api-key ", cur_api_key)
	yts, err := youtube.NewService(ctx, option.WithAPIKey(cur_api_key))
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

func PrimitiveDateToUtcString(date_time primitive.DateTime) string {
	return date_time.Time().UTC().Format(time.RFC3339)
}

func fetchVideoAndUpdateDB() {
	if cur_api_key == "" {
		updateApiKeyAndIndex()
	}
	youtube_data, err := fetchFromYoutube(YOUTUBE_QUERY, YOUTUBE_QUERY, publish_time, next_page_token)
	if err != nil {
		fmt.Println(err)
		updateApiKeyAndIndex()
		return
	}
	fmt.Println("--youtube-data----", youtube_data.NextPageToken, " ", next_page_token)
	next_page_token = youtube_data.NextPageToken
	fmt.Println("youtube_data length ", len(youtube_data.Items))
	models := make([]mongo.WriteModel, 0)
	for _, item := range youtube_data.Items {
		t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			fmt.Println(err)
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
	resp, err := database.InsertManyItemToDB(models)
	if err != nil {
		fmt.Println("Error Inserting data ", err)
	}
	fmt.Println("Inserted successfully ", resp)
}

func UpdateLatestVideos() {
	ticker := time.NewTicker(20 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				fetchVideoAndUpdateDB()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
