package model

import (
	"time"
)

type RentDetail struct {
	ID        uint      `json:"id"`
	RentID    uint      `json:"rent_id"`
	ItemID    uint      `json:"item_id"`
	Quantity  int       `json:"quantity"`
	Price     int       `json:"price"`
	DateStart time.Time `json:"date_start"`
	DateEnd   time.Time `json:"date_end"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}