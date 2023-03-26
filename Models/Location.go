package model

type Location struct {
	State string `bson:"state"`
	City  string `bson:"city"`
}