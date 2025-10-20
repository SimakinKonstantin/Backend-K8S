package models

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Program struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Price        int           `json:"price"`
	WasConfirmed string        `json:"wasconfirmed"`
	ConfirmedBy  bson.ObjectID `json:"confirmedby"`
}

type Error struct {
	Message string
}

// Структура для redis.
type CachedProgram struct {
	Id           bson.ObjectID `bson:"_id" json:"_id"`
	Name         string        `bson:"name" json:"name"`
	Description  string        `bson:"description" json:"description"`
	Price        int           `bson:"price" json:"price"`
	WasConfirmed string        `bson:"wasconfirmed" json:"wasconfirmed"`
	ConfirmedBy  bson.ObjectID `bson:"confirmedby" json:"confirmedby"`
}

// Реализация интерфейса, необходимого для работы с redis.
func (cp CachedProgram) MarshalBinary() ([]byte, error) {
	return json.Marshal(cp)
}
