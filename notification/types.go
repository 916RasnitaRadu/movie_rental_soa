package main

type NotificationEvent struct {
	Type    string `json:"type"`
	Email   string `json:"email"`
	Message string `json:"message"`
}
