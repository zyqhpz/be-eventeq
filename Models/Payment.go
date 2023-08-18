package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
	ID        	primitive.ObjectID 		`bson:"_id,omitempty"`
	BookingID 	primitive.ObjectID 		`bson:"booking_id"`
	BillCode  	string             		`bson:"bill_code"`
	Status    	int32              		`bson:"status"`
	CreatedAt 	time.Time             	`bson:"created_at"`
	UpdatedAt 	time.Time             	`bson:"updated_at"`
}