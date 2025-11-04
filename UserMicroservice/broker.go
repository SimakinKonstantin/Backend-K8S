package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"log/slog"
	"time"
)

func ProcessMessage(consumer *kafka.Reader, producer *kafka.Writer, db *DbProcessor) error {
	for {
		msg, err := consumer.ReadMessage(context.TODO())
		if err != nil {
			return err
		}

		//slog.Info(fmt.Sprintf("Users::прочитано сообщение из очереди ObjectId: %s", msg.Key))

		// Увеличиваем счетчик в базе.
		err = db.IncreaseCounter()
		if err != nil {
			return err
		}

		// Получаем время обновления.
		var updatedStatus string
		bsonId, err := bson.ObjectIDFromHex(string(msg.Value))
		if err != nil {
			slog.Error(err.Error())
		}

		if db.FindUser(bsonId) {
			updatedStatus = time.Now().String()

		} else {
			updatedStatus = "Not accepted. Invlid user id"
		}

		producer.WriteMessages(context.TODO(), kafka.Message{
			Key:   msg.Key,
			Value: []byte(updatedStatus),
		})

		//slog.Info(fmt.Sprintf("Users::записано сообщение в очередь ObjectId: %s", msg.Key))
	}
	return nil
}
