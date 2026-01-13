package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishRabbitNotification(r RentalRequest) {
	channel, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Printf("ERROR: rabbit dial in publish %v\n", err)
		return
	}
	defer channel.Close()

	conn, err := channel.Channel()
	if err != nil {
		log.Printf("ERROR: rabbit channel: %v\n", err)
		return
	}
	defer conn.Close()

	queue, err := conn.QueueDeclare("email_notifications", true, false, false, false, nil)
	if err != nil {
		log.Printf("ERROR: queue declare: %v\n", err)
		return
	}

	payload := map[string]interface{}{
		"type": "rental.created",
		"data": r,
	}

	log.Println("INFO: sending payload ", payload)
	bs, _ := json.Marshal(payload)
	if err := conn.Publish("", queue.Name, false, false, amqp.Publishing{ContentType: "application/json", Body: bs}); err != nil {
		log.Printf("ERROR: rabbit publish: %v\n", err)
		return
	}
}
