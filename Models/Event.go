package model

import (
	"time"
)

type Event struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
	DateStart   time.Time `json:"date_start"`
	DateEnd     time.Time `json:"date_end"`
	OrganizedBy int    `json:"organized_by"`
}