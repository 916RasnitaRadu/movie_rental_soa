package main

type Movie struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Genre  string `json:"genre"`
	Length int64  `json:"length"`
}
