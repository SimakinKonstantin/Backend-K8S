package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	_ "lab1/docs"
	"log/slog"
	"net/http"
	"os"
)

// @title Swagger Example API
// @version 2.0
// @description Симакин К.С. Акчурин А.Р.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
func main() {
	slog.Info("7 Вариант. Программа спортивных тренировок")

	db := dbProcessor{Client: nil, DbName: "training_programs", ColName: "programs"}
	db.Client, _ = mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_URL")))

	defer func() {
		if err := db.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR")})
	defer rdb.Close()

	Producer := kafka.Writer{Addr: kafka.TCP(os.Getenv("KAFKA_ADDR")), Topic: "post"}
	defer Producer.Close()

	Consumer := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{os.Getenv("KAFKA_ADDR")}, Topic: "request"})
	defer Consumer.Close()

	app := App{Database: &db, Redis: rdb, Producer: &Producer}

	server := &http.Server{
		Addr:    ":8081",
		Handler: app.routes(),
	}

	// Просушивание кафка очереди.
	go Listen(Consumer, &app)

	server.ListenAndServe()
}
