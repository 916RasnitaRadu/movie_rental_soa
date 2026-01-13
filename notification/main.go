package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitUrl string

func main() {
	rabbitUrl = getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Fatalf("ERROR: rabbit dial %v\n", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("ERROR: channel err %v\n", err)
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"email_notifications",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("ERROR: queue declaration %v\n", err)
	}

	messages, err := channel.Consume(
		queue.Name,
		"",    // consumer
		false, //auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("ERROR: messages consumer %v\n", err)
	}

	log.Println("INFO: notification service waiting for messages...")

	// Handle messages in a backgrnd goroutine
	go func() {
		for d := range messages {
			var eventMap map[string]interface{}
			if err := json.Unmarshal(d.Body, &eventMap); err != nil {
				log.Printf("ERROR: invalid message %v\n", err)
				_ = d.Nack(false, false) // discard invalid message
				continue
			}

			log.Printf("INFO: received message for %s\n", eventMap)

			// call OpenFaaS func
			if err := triggerEmailFaaS("data"); err != nil {
				log.Printf("ERROR: calling FaaS %v\n", err)
				_ = d.Ack(false)
				continue
			}

			_ = d.Ack(false)
		}
	}()

	// wait for termination signal (CTRL + C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("INFO: shutting down notification service")
}

// calls OpenFaas for sending emails
func triggerEmailFaaS(body string) error {
	gatewayURL := "http://host.docker.internal:31112/function/send-email"

	payload := map[string]string{
		"body": body,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: failed to marshal email request %w\n", err)
		return fmt.Errorf("ERROR: failed to marshal email request %w", err)
	}

	response, err := http.Post(gatewayURL, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		log.Printf("ERROR: failed calling send-email %w\n", err)
		return fmt.Errorf("ERROR: failed calling send-email %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("ERROR:send-email failed with status %d\n", response.StatusCode)
		return fmt.Errorf("ERROR: send-email failed with status %d", response.StatusCode)
	}

	return nil
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
