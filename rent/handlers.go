package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func HandleGetRent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var re Rental
	var rentalTime, expiresAt time.Time
	err := db.QueryRow("SELECT id, movie_id, username, status, created_at, expires_at FROM rental WHERE id = ?", id).
		Scan(&re.ID, &re.MovieID, &re.User, &re.Status, &rentalTime, &expiresAt)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	re.RentTime = rentalTime
	re.ExpiresAt = expiresAt
	json.NewEncoder(w).Encode(re)
}

func HandleListRents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, movie_id, username, status, created_at, expires_at FROM rental")
	if err != nil {
		log.Println("ERROR: error when retrieving the rentals")
		http.Error(w, "error when retrieving the rentals", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Rental{}
	for rows.Next() {
		var r Rental
		var rentalTime, expiresAt time.Time
		if err := rows.Scan(&r.ID, &r.MovieID, &r.User, &r.Status, &rentalTime, &expiresAt); err != nil {
			log.Println("ERROR: error when converting rental ", err)
			continue
		}
		r.RentTime = rentalTime
		r.ExpiresAt = expiresAt
		out = append(out, r)
	}
	json.NewEncoder(w).Encode(out)
}

func HandleCreateRent(w http.ResponseWriter, r *http.Request) {
	var rent RentalRequest
	if err := json.NewDecoder(r.Body).Decode(&rent); err != nil {
		log.Println("ERROR: bad request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := db.Exec("INSERT INTO rental (movie_id, username, status) VALUES (?, ?, ?)",
		rent.MovieID, rent.User, rent.Status); err != nil {
		log.Println("ERROR: error at creating new rental ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// publish to RabbitMQ queue for email notifications
	go PublishRabbitNotification(rent)

	// Optional: also write to kafka rentals topic
	// go produceKafkaRental(rent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rent)
}
