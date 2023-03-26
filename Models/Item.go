package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
	Name        string 				`bson:"name"`
	Description string 				`bson:"description"`
	Category    int32				`bson:"type"`
	Price       float32    			`bson:"price"`
	Stock       int32    			`bson:"stock"`

	// Relationship OwnedBy to User id
	OwnedBy 	primitive.ObjectID	`bson:"owned_by"`

	// Timestamp
	CreatedAt 	time.Time 			`bson:"created_at"`
	UpdatedAt 	time.Time 			`bson:"updated_at"`

	// Image
	Image 		[]string 			`bson:"image"`

	/* Status
		0 = Disabled
		1 = Available
		2 = Rented
		3 = Reserved
	*/
	Status 		int32 				`bson:"status"`
}