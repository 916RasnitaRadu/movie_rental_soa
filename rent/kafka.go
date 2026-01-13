package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// kafka consumer that listens to movie.created / movie.deleted events
func ConsumeMovieEvents() {
	for {
		message, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("ERROR: kafka read error: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var ev map[string]interface{}
		if err := json.Unmarshal(message.Value, &ev); err != nil {
			log.Printf("ERROR: error when unmarshalling json: %v\n", err)
			continue
		}

		event, _ := ev["type"].(string)
		data := ev["data"]
		if event == "movie.created" {
			// data is a movie object
			bs, _ := json.Marshal(data)
			var m MovieCache

			_ = json.Unmarshal(bs, &m)
			_, _ = db.Exec("REPLACE INTO movies_cache (id, name, genre, length_minutes, last_seen) VALUES (?, ?, ?, ?, NOW())",
				m.ID, m.Name, m.Genre, m.Length)

			log.Println("INFO: cached movie ", m.ID, m.Name)
		} else if event == "movie.deleted" {
			bs, _ := json.Marshal(data)
			var dd struct {
				ID int64 `json:"id"`
			}

			_ = json.Unmarshal(bs, &dd)
			_, _ = db.Exec("DELETE FROM movies_cache WHERE id = ?", dd.ID)

			log.Println("INFO: removed movie from cache: ", dd.ID)
		}
	}
}

func ProduceKafkaRental(r Rental) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, _ := json.Marshal(r)
	if err := kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(int(r.ID))),
		Value: val,
	}); err != nil {
		log.Println("ERROR: failed to write kafka message:", err)
	}
}
