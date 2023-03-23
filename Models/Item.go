package model

import (
	"time"
)

type Item struct {
	ID          uint   `bson:"id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
	Type        int    `bson:"type"`
	Price       int    `bson:"price"`
	Stock       int    `bson:"stock"`

	// Relationship OwnedBy to User id
	OwnedBy 	int `bson:"owned_by"`

	// Timestamp
	CreatedAt 	time.Time `bson:"created_at"`
	UpdatedAt 	time.Time `bson:"updated_at"`

	// Image
	Image 		string `bson:"image"`

	// Status
	Status 		int `bson:"status"`
}