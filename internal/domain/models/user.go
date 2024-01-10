package models

import (
	"time"
)

type User struct {
	ID          int64     `json:"ID"`
	FirstName   string    `json:"firstName"`
	SecondName  string    `json:"secondName"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	Activated   bool      `json:"activated"`
	CreatedAt   time.Time `json:"created_at"`
	Roles       []string  `json:"roles"`
	Barcode     string    `json:"barcode"`
	PhoneNumber string    `json:"phoneNumber"`
	Major       string    `json:"major"`
	Group       string    `json:"group"`
	Year        int       `json:"year"`
}
