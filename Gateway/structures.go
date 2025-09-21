package main

import "go.mongodb.org/mongo-driver/v2/bson"

type Program struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Price        int           `json:"price"`
	WasConfirmed string        `json:"wasconfirmed"`
	ConfirmedBy  bson.ObjectID `json:"confirmedby"`
}

type User struct {
	Login   string `json:"login"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     int    `json:"age"`
}

type Error struct {
	Message string
}
