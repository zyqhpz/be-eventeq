package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID        primitive.ObjectID 	`bson:"_id,omitempty"`
	Name        string 				`bson:"name"`
	Description string 				`bson:"description"`
	Location	Location			`bson:"location"`
	DateStart   time.Time 			`bson:"date_start"`
	DateEnd     time.Time 			`bson:"date_end"`
	OrganizedBy primitive.ObjectID	`bson:"organized_by"`
	CreatedAt 	time.Time 			`bson:"created_at"`
	UpdatedAt 	time.Time 			`bson:"updated_at"`
}