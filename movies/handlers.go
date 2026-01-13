package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func HandleGetMovie(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var movie Movie
	err := db.QueryRow("SELECT id, name, genre, length_minutes FROM movie WHERE id = ?", id).Scan(&movie.ID, &movie.Name, &movie.Genre, &movie.Length)
	if err != nil {
		log.Println("INFO: movie not found")
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

func HandleListMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, genre, length_minutes FROM movie")
	if err != nil {
		log.Println("ERROR: error when retrieving the movies")
		http.Error(w, "error when retrieving the movies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Name, &m.Genre, &m.Length); err != nil {
			log.Println("ERROR: error when converting movie ", err)
			continue
		}
		out = append(out, m)
	}

	json.NewEncoder(w).Encode(out)
}

func HandleCreateMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Println("ERROR: bad request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := db.Exec("INSERT INTO movie(name, genre, length_minutes) VALUES (?, ?, ?)", m.Name, m.Genre, m.Length); err != nil {
		log.Println("ERROR: error at creating new movie ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// publish kafka event: movie.created
	event := map[string]interface{}{
		"type": "movie.created",
		"data": m,
	}

	b, _ := json.Marshal(event)
	_ = kWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(strconv.Itoa(int(m.ID))),
		Value: b,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func HandleDeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if _, err := db.Exec("DELETE FROM movie WHERE id = ?", id); err != nil {
		log.Printf("ERROR: error when deleting the movie with id %v.\n", id)
		http.Error(w, "error when deleting the movie", http.StatusInternalServerError)
		return
	}

	// publish kafka event: movie.deleted
	event := map[string]interface{}{
		"type": "movie.deleted",
		"data": map[string]string{"id": id},
	}

	b, _ := json.Marshal(event)
	_ = kWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(id),
		Value: b,
	})

	w.WriteHeader(http.StatusNoContent)
}
