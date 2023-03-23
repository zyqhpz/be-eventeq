package model

import (
	"time"
)

type RentDetail struct {
	ID        uint      `bson:"id"`
	RentID    uint      `bson:"rent_id"`
	ItemID    uint      `bson:"item_id"`
	Quantity  int       `bson:"quantity"`
	Price     int       `bson:"price"`
	DateStart time.Time `bson:"date_start"`
	DateEnd   time.Time `bson:"date_end"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}