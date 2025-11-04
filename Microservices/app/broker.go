package app

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"log/slog"
)

// userId - id пользователя, который подтверждает; objectId - id объекта для подтверждения
func Write(producer *kafka.Writer, objectId string, userId string) error {
	//slog.Info(fmt.Sprintf("Object::Записано сообщение в очередь ObjectId: %s", objectId))

	return producer.WriteMessages(context.TODO(), kafka.Message{
		Key:   []byte(objectId),
		Value: []byte(userId),
	})
}

func Listen(consumer *kafka.Reader, app *App) {
	for {
		msg, err := consumer.ReadMessage(context.TODO())

		if err != nil {
			slog.Error(err.Error())
			return
		}

		//slog.Info(fmt.Sprintf("Object::Получено сообщение из очереди ObjectId: %s", string(msg.Key)))

		objectId, err := bson.ObjectIDFromHex(string(msg.Key))
		if err != nil {
			slog.Error(err.Error())
			return
		}
		status := string(msg.Value)

		err = app.UpdateStatus(objectId, status)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}
}
