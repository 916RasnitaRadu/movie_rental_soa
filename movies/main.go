package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

var (
	db          *sql.DB
	kafkaBroker string
	kWriter     *kafka.Writer
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

	kafkaBroker = getEnv("KAFKA_BROKER", "localhost:9092")
	kWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "movies",
		Balancer: &kafka.LeastBytes{},
	})

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Use(jwtAuthMiddleware)
	api.HandleFunc("/movies", HandleListMovies).Methods("GET")
	api.HandleFunc("/movies", HandleCreateMovie).Methods("POST")
	api.HandleFunc("/movies/{id}", HandleGetMovie).Methods("GET")
	api.HandleFunc("/movies/{id}", HandleDeleteMovie).Methods("DELETE")

	port := getEnv("PORT", "8001")
	log.Printf("movies service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
