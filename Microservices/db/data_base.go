package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"lab1/graphq/graph/model"
	"lab1/models"
)

type Processor struct {
	Client  *mongo.Client
	DbName  string
	ColName string
}

func (db Processor) GetAllPrograms() ([]bson.D, error) {
	cursor, err := db.Client.Database(db.DbName).Collection(db.ColName).Find(context.TODO(), bson.D{})

	// Парсинг результата.
	var programs []bson.D
	cursor.All(context.TODO(), &programs)
	return programs, err
}

func (db Processor) GetProgram(programID string) (bson.D, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	var program bson.D
	id, _ := bson.ObjectIDFromHex(programID)
	err := collection.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&program)
	return program, err
}

func (db Processor) FilterPrograms(filter model.ProgramFilterInput) ([]models.CachedProgram, error) {
	dbFilter := bson.D{}
	if filter.Name != nil {
		dbFilter = append(dbFilter, bson.E{"name", *filter.Name})
	}

	if filter.Description != nil {
		dbFilter = append(dbFilter, bson.E{"description", *filter.Description})
	}

	if filter.Price != nil {
		dbFilter = append(dbFilter, bson.E{"price", *filter.Price})
	}

	if filter.Wasconfirmed != nil {
		dbFilter = append(dbFilter, bson.E{"wasconfirmed", *filter.Wasconfirmed})
	}

	if filter.ConfirmedBy != nil {
		dbFilter = append(dbFilter, bson.E{"confirmedby", *filter.ConfirmedBy})
	}

	cursor, err := db.Client.Database(db.DbName).Collection(db.ColName).Find(context.TODO(), dbFilter)

	// Парсинг результата.
	var programs []models.CachedProgram
	cursor.All(context.TODO(), &programs)
	return programs, err
}

func (db Processor) AddProgram(newProgram *models.Program) (*mongo.InsertOneResult, error) {
	result, err := db.Client.Database(db.DbName).Collection(db.ColName).InsertOne(context.TODO(), newProgram)
	return result, err
}

func (db Processor) UpdateProgram(programID string, newValues bson.D) error {
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

func (db Processor) DeleteProgram(programID string) error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	id, _ := bson.ObjectIDFromHex(programID)
	res, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", id}})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db Processor) DeleteAllPrograms() error {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.DeleteMany(context.TODO(), bson.D{})

	if res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

func (db Processor) CountPrograms() (int64, error) {
	collection := db.Client.Database(db.DbName).Collection(db.ColName)
	res, err := collection.CountDocuments(context.TODO(), bson.D{})
	return res, err
}
