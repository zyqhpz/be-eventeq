package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
	UserID 		primitive.ObjectID 	`bson:"user_id"`
	OwnerID		primitive.ObjectID 	`bson:"owner_id"`
	Items 		[]Item 				`bson:"items"`
	StartDate 	string 				`bson:"start_date"`
	EndDate 	string 				`bson:"end_date"`
	SubTotal 	float64 			`bson:"sub_total"`
	ServiceFee 	float64 			`bson:"service_fee"`
	GrandTotal 	float64 			`bson:"grand_total"`
	Status 		int32 				`bson:"status"`
	CreatedAt 	time.Time 			`bson:"created_at"`
	UpdatedAt 	time.Time 			`bson:"updated_at"`
}