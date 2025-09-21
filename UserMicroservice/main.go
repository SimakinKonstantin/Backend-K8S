package main

import (
	_ "UserMicroservice/docs"
	"context"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log/slog"
	"net/http"
	"os"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @title Swagger Example API
// @version 2.0
// @description Симакин К.С. Акчурин А.Р.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
func main() {

	db := DbProcessor{Client: nil, DbName: "training_programs", ColName: "users"}

	var err error
	db.Client, err = mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		slog.Error("Ошибка подключения к БД: ", err)
		return

	}

	if err = db.Client.Ping(context.TODO(), nil); err != nil {
		slog.Info("Ошибка пинга БД: ", err)
		return
	}

	defer func() {
		if err := db.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	app := App{Database: &db}

	server := &http.Server{
		Addr:    ":8080",
		Handler: app.Routes(),
	}

	Producer := kafka.Writer{Addr: kafka.TCP(os.Getenv("KAFKA")), Topic: "request"}
	defer Producer.Close()

	Consumer := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{os.Getenv("KAFKA")}, Topic: "post"})
	defer Consumer.Close()

	// Параллельно принимаем обрабатываем очередь.
	go ProcessMessage(Consumer, &Producer, &db)

	err = server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
