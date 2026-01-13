package main

import "time"

type Rental struct {
	ID        int64     `json:"id"`
	MovieID   int64     `json:"movie_id"`
	User      string    `json:"user"`
	RentTime  time.Time `json:"rent_time"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RentalRequest struct {
	MovieID int64  `json:"movie_id"`
	User    string `json:"user"`
	Status  string `json:"status"`
}

type MovieCache struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Genre  string `json:"genre"`
	Length int64  `json:"length"`
}
