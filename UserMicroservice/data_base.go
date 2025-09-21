package main

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
)

// Информация о базе данных.
type DbProcessor struct {
	Client  *mongo.Client
	DbName  string
	ColName string
}

// Информация о таблице БД.
type User struct {
	Login   string `json:"login"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     int    `json:"age"`
}

func (db DbProcessor) GetAllUsers() ([]bson.D, error) {
	cursor, err := db.Client.Database(db.DbName).Collection(db.ColName).Find(context.TODO(), bson.D{})
	if err != nil {
		slog.Error("GetAllUsers err:", err)
		return nil, err
	}

	// Парсинг результата.
	var users []bson.D
	cursor.All(context.TODO(), &users)
	return users, err
}

func (db DbProcessor) GetUser(UsesrId string) (bson.D, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	var user bson.D
	id, _ := bson.ObjectIDFromHex(UsesrId)
	err := collection.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&user)
	return user, err
}

func (db DbProcessor) AddUser(newUser *User) (*mongo.InsertOneResult, error) {
	result, err := db.Client.Database(db.DbName).Collection(db.ColName).InsertOne(context.TODO(), newUser)
	return result, err
}

func (db DbProcessor) UpdateUser(userID string, newValues bson.D) error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	id, _ := bson.ObjectIDFromHex(userID)
	res, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", newValues}})
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db DbProcessor) DeleteUser(userID string) error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	id, _ := bson.ObjectIDFromHex(userID)
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", id}})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db DbProcessor) DeleteAllUsers() error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.DeleteMany(context.TODO(), bson.D{})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db DbProcessor) CountUsers() (int64, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.CountDocuments(context.TODO(), bson.D{})
	return res, err
}

func (db DbProcessor) IncreaseCounter() error {

	col := db.Client.Database(db.DbName).Collection("requestCounter")

	_, err := col.UpdateOne(context.TODO(), bson.D{{"_id", 1}}, bson.D{{"$inc", bson.D{{"Value", 1}}}})
	if err != nil {
		return AppError{"ошибка обновления счетчика"}
	}

	return nil
}

func (db *DbProcessor) FindUser(userId bson.ObjectID) bool {
	col := db.Client.Database(db.DbName).Collection("users")

	var result bson.M
	err := col.FindOne(context.TODO(), bson.D{{"_id", userId}}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false
		}
		slog.Error("ошибка в db.FindUser")
		return false
	}
	return true
}

func (db *DbProcessor) FindUserByLogin(login string) bool {
	col := db.Client.Database(db.DbName).Collection("users")
	var result bson.M
	err := col.FindOne(context.TODO(), bson.D{{"login", login}}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false
		}
		slog.Error("ошибка в db.FindUser")
		return false
	}
	return true
}
