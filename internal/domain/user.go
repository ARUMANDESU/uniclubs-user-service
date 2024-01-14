package domain

import "time"

type User struct {
	ID          int64     `json:"ID"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"secondName"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	Role        string    `json:"role"`
	Barcode     string    `json:"barcode"`
	PhoneNumber string    `json:"phoneNumber"`
	Major       string    `json:"major"`
	Group       string    `json:"group"`
	Year        int       `json:"year"`
}
