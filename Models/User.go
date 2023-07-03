package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        			primitive.ObjectID 	`bson:"_id,omitempty"`
	FirstName 			string             	`bson:"first_name"`
	LastName  			string             	`bson:"last_name"`
	Password  			string             	`bson:"password"`
	Email     			string             	`bson:"email"`
	NoPhone   			string             	`bson:"no_phone"`

	IsAvatarImageSet 	bool 				`bson:"isAvatarImageSet"`
	ProfileImage     	primitive.ObjectID	`bson:"profile_image"`

	Location  			Location           	`bson:"location"`

	// Timestamp
	CreatedAt 			time.Time 			`bson:"created_at"`
	UpdatedAt 			time.Time 			`bson:"updated_at"`
}