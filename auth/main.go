package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var (
	db     *sql.DB
	jwtKey []byte
)

func main() {
	dsn := getEnv("MYSQL_CONN_STRING", "root:rootpass@tcp(mysql:3306)/microdb?parseTime=true")

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetMaxOpenConns(10)
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	jwtKey = []byte(getEnv("JWT_SECRET", "supersecret"))

	r := mux.NewRouter()
	r.HandleFunc("/register", HandleRegister).Methods("POST")
	r.HandleFunc("/login", HandleLogin).Methods("POST")

	port := getEnv("PORT", "8000")
	log.Printf("auth service listening on: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
