package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID        		primitive.ObjectID 	`bson:"_id,omitempty"`
	Name        	string 				`bson:"name"`
	Description 	string 				`bson:"description"`
	Location		Location			`bson:"location"`
	StartDate   	string 				`bson:"start_date"`
	EndDate     	string 				`bson:"end_date"`
	// Category		string 				`bson:"category"`

	/* 0: Found, 1: Finding */
	Status	  		int 				`bson:"status"`
	OrganizedBy 	primitive.ObjectID	`bson:"organized_by"`
	CreatedAt 		time.Time 			`bson:"created_at"`
	UpdatedAt 		time.Time 			`bson:"updated_at"`
}