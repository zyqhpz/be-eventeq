package model

import (
	"time"
)

type Event struct {
	ID          uint   		`bson:"id"`
	Name        string 		`bson:"name"`
	Description string 		`bson:"description"`
	State       string 		`bson:"state"`
	DateStart   time.Time 	`bson:"date_start"`
	DateEnd     time.Time 	`bson:"date_end"`
	OrganizedBy int    		`bson:"organized_by"`
}