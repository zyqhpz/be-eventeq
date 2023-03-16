package model

import (
	"time"
)

type Item struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`

	// Relationship
	OwnedBy User `json:"owned_by" gorm:"foreignKey:UserID"`

	// Timestamp
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Image
	Image string `json:"image"`

	// Status
	Status int `json:"status"`
}