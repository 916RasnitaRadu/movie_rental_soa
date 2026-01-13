package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
)

var (
	db          *sql.DB
	amqpConn    *amqp.Connection
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
	rabbitUrl   string
)

func main() {
	dsn := getEnv("MYSQL_CONN_STRING", "root:rootpass@tcp(mysql:3306)/microdb?parseTime=true")

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	rabbitUrl = getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	amqpConn, err = amqp.Dial(rabbitUrl)
	if err != nil {
		log.Fatalf("rabbit dial: %v", err)
	}

	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "movies",
		GroupID:  "rent-movies-consumer",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "movies",
	})

	// start backgrnd consumer for movie events
	go ConsumeMovieEvents()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Use(jwtAuthMiddleware)
	api.HandleFunc("/rent/{id}", HandleGetRent).Methods("GET")
	api.HandleFunc("/rent", HandleListRents).Methods("GET")
	api.HandleFunc("/rent", HandleCreateRent).Methods("POST")

	port := getEnv("PORT", "8002")
	log.Printf("rent service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
