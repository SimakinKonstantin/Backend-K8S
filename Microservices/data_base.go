package main

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type dbProcessor struct {
	Client  *mongo.Client
	DbName  string
	ColName string
}

func (db dbProcessor) getAllPrograms() ([]bson.D, error) {
	cursor, err := db.Client.Database(db.DbName).Collection(db.ColName).Find(context.TODO(), bson.D{})

	// Парсинг результата.
	var programs []bson.D
	cursor.All(context.TODO(), &programs)
	return programs, err
}

func (db dbProcessor) getProgram(programID string) (bson.D, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	var program bson.D
	id, _ := bson.ObjectIDFromHex(programID)
	err := collection.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&program)
	return program, err
}

func (db dbProcessor) addProgram(newProgram *Program) (*mongo.InsertOneResult, error) {
	result, err := db.Client.Database(db.DbName).Collection(db.ColName).InsertOne(context.TODO(), newProgram)
	return result, err
}

func (db dbProcessor) updateProgram(programID string, newValues bson.D) error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	id, _ := bson.ObjectIDFromHex(programID)
	res, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", newValues}})
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db dbProcessor) deleteProgram(programID string) error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	id, _ := bson.ObjectIDFromHex(programID)
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", id}})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db dbProcessor) deleteAllPrograms() error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.DeleteMany(context.TODO(), bson.D{})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db dbProcessor) countPrograms() (int64, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.CountDocuments(context.TODO(), bson.D{})
	return res, err
}
