package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID 	`bson:"_id,omitempty"`
	FirstName string             	`bson:"first_name"`
	LastName  string             	`bson:"last_name"`
	Username  string             	`bson:"username"`
	Password  string             	`bson:"password"`
	Email     string             	`bson:"email"`

	// Timestamp
	CreatedAt time.Time 			`bson:"created_at"`
	UpdatedAt time.Time 			`bson:"updated_at"`
}