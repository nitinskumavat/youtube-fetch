package database

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Video struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	PublishedAt primitive.DateTime `json:"published_at" bson:"published_at"`
	Thumbnail   string             `json:"thumbnail" bson:"thumbnail"`
	VideoETag   string             `json:"video_etag" bson:"video_etag"`
}

var collection *mongo.Collection

func ConnectToDB() *mongo.Collection {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://nitin:nitin123.@videocluster.8coh86y.mongodb.net/test"))
	if err != nil {
		log.Fatal("Error creating mongodb client ", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error connecting to mogodb ", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Conneted to DB")
	collection = client.Database("youtube-videos").Collection("videos")
	return collection
}

func InsertManyItemToDB(items []mongo.WriteModel) (mongo.BulkWriteResult, error) {
	opts := options.BulkWrite()
	opts.SetOrdered(false)
	res, err := collection.BulkWrite(context.TODO(), items, opts)
	if err != nil {
		fmt.Println("err")
		return mongo.BulkWriteResult{}, err
	}

	fmt.Printf(
		"inserted %v and deleted %v documents\n", res.InsertedCount, res.DeletedCount)
	return *res, nil
}

func GetQueryVideos(c *gin.Context) {
	// filter := bson.M{}
	query_string := c.Query("query")
	fmt.Println("query string ", query_string)
	search_stage := mongo.Pipeline{
		bson.D{
			{Key: "$search", Value: bson.D{
				{Key: "text", Value: bson.D{
					{Key: "path", Value: []string{"title", "description"}},
					{Key: "query", Value: query_string},
					{Key: "fuzzy", Value: bson.D{{Key: "maxEdits", Value: 2}}},
				}},
			}},
		},
	}
	dbctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	opts := options.Aggregate().SetMaxTime(5 * time.Second)
	// opts.SetHint()
	cursor, err := collection.Aggregate(dbctx, search_stage, opts)
	fmt.Println(cursor.Current.String())
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(dbctx)
	videoList := make([]Video, 0)
	for cursor.Next(dbctx) {
		video := &Video{}
		err := cursor.Decode(&video)
		if err != nil {
			log.Fatal(err)
		}
		videoList = append(videoList, *video)
	}
	c.JSON(http.StatusOK, gin.H{"data": videoList})
}

func GetVideos(c *gin.Context) {
	next_key := c.Query("next_key")
	dbctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	opts := options.Find()
	if next_key != "" {
		next_id, err := primitive.ObjectIDFromHex(next_key)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Invlaid next key"})
		}
		opts.SetHint(bson.D{{Key: "_id", Value: 1}})
		opts.SetMin(bson.D{{Key: "_id", Value: next_id}})
	}
	opts.SetSort(bson.D{{Key: "published_at", Value: -1}})
	opts.SetLimit(11)

	fmt.Println("----filter ---- ", opts)
	cursor, err := collection.Find(dbctx, bson.M{}, opts)
	if err != nil {
		log.Fatal(err)
	}
	videoList := make([]Video, 0)
	item_count := 0
	next_key = ""
	for cursor.Next(dbctx) {
		item_count += 1
		video := &Video{}
		err := cursor.Decode(&video)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(video.PublishedAt.Time().UTC().Format(time.RFC3339))
		if item_count > 10 {
			next_key = video.Id.Hex()
			break
		}
		videoList = append(videoList, *video)
	}
	c.JSON(http.StatusOK, gin.H{"data": videoList, "next_key": next_key})
}

func DeleteMany(c *gin.Context) {
	dbctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	res, _ := collection.DeleteMany(dbctx, bson.M{})
	c.JSON(http.StatusOK, gin.H{"res": res})
}

func GetTopRow() Video {
	dbctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	video := &Video{}
	cursor := collection.FindOne(dbctx, bson.M{})
	if err := cursor.Decode(&video); err != nil {
		fmt.Println("Error reading top document ", err)
		return *video
	}
	return *video
}
