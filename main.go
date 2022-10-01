package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
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
