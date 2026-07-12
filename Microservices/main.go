package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/vektah/gqlparser/v2/ast"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"lab1/app"
	"lab1/db"
	_ "lab1/docs"
	"lab1/graphq/graph"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// @title Swagger Example API
// @version 2.0
// @description Сервис программ тренировок
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host
func main() {
	slog.Info("7 Вариант. Программа спортивных тренировок")
	database := db.Processor{Client: nil, DbName: "training_programs", ColName: "programs"}
	database.Client, _ = mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_URL")))

	defer func() {
		if err := database.Client.Disconnect(context.TODO()); err != nil {
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

	application := app.App{Database: &database, Redis: rdb, Producer: &Producer}

	server := &http.Server{
		Addr:    ":8081",
		Handler: application.Routes(),
	}

	// Просушивание кафка очереди.
	go app.Listen(Consumer, &application)

	go startGQL(&application)

	server.ListenAndServe()
}

func startGQL(application *app.App) {
	const graphQLPort = "8010"

	port := graphQLPort

	resolver := graph.NewResolver(application)
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
