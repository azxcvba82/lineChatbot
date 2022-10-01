package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserMessage struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func main() {

	// config
	viper.SetConfigFile("./conf/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()

	bot, err := linebot.New(
		viper.GetString("channelSecret"),
		viper.GetString("channelToken"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// mongo
	uri := viper.GetString("mongoDBConnectionString")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		panic(fmt.Errorf("Fatal error mongo: %s \n", err))
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(fmt.Errorf("Fatal error mongo disconnect: %s \n", err))
		}
	}()

	// gin
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/save", func(c *gin.Context) {

		var jsonData UserMessage
		err := c.BindJSON(&jsonData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		coll := client.Database("demo").Collection("line")

		doc := bson.D{{"user", jsonData.User}, {"message", jsonData.Message}}
		result, err := coll.InsertOne(context.TODO(), doc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	})

	r.POST("/query", func(c *gin.Context) {

		var jsonData UserMessage
		err := c.BindJSON(&jsonData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		coll := client.Database("demo").Collection("line")

		cursor, err := coll.Find(context.TODO(), bson.D{{"user", jsonData.User}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		var userMessages []UserMessage
		for _, result := range results {
			output, err := json.Marshal(result)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
			}
			var userMessage UserMessage
			err = json.Unmarshal(output, &userMessage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
			}

			userMessages = append(userMessages, userMessage)
		}

		c.JSON(http.StatusOK, gin.H{
			"result": userMessages,
		})
	})

	r.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)

		if err != nil {
			if err == linebot.ErrInvalidSignature {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "ErrInvalidSignature",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "500: StatusInternalServerError",
				})
			}
			return
		}

		var messageResponse string
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					fmt.Printf("%v", message)
					messageResponse = message.Text

					coll := client.Database("demo").Collection("line")

					doc := bson.D{{"user", event.Source.UserID}, {"message", message.Text}}
					_, err := coll.InsertOne(context.TODO(), doc)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"message": err.Error(),
						})
					}

				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": messageResponse,
		})
	})

	r.Run(viper.GetString("port")) // listen and serve on 0.0.0.0:80 (for windows "localhost:80")
}
