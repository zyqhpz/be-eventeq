package model

type Location struct {
	State    string `bson:"state"`
	District string `bson:"district"`
}