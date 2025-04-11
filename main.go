package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

// Define your data model
type IoTData struct {
	DeviceID    string  `json:"device_id"`
	ApiKey      string  `json:"api_key"`
	Temperature float64 `json:"temperature"`
	Speed       float64 `json:"speed"`
}

type ChargeData struct {
	DeviceID string  `json:"device_id"`
	ApiKey   string  `json:"api_key"`
	Voltage  float64 `json:"voltage"`
	Current  float64 `json:"current"`
}

var rabbitChannel *amqp.Channel

func initRabbitMQ() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open a channel: %s", err)
	}

	// Declare both queues
	queues := []string{"iot-data", "iot-data-charge"}
	for _, q := range queues {
		_, err = ch.QueueDeclare(
			q,     // queue name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Fatalf("❌ Failed to declare queue %s: %s", q, err)
		}
	}

	log.Println("✅ Connected to RabbitMQ and queues declared.")
	return ch
}

func iotCarHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var data IoTData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	err = rabbitChannel.Publish(
		"",         // exchange
		"iot-data", // routing key (queue name)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("❌ Failed to publish message: %s", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	log.Printf("📤 Sent message to RabbitMQ: %s", string(body))
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Data received and queued")
}

func iotChargeHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var data ChargeData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	err = rabbitChannel.Publish(
		"",                // exchange
		"iot-data-charge", // queue
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("❌ Failed to publish charge message: %s", err)
		http.Error(w, "Failed to send charge message", http.StatusInternalServerError)
		return
	}

	log.Printf("📤 Sent charge data to RabbitMQ: %s", string(body))
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Charge data received and queued")
}

func main() {
	rabbitChannel = initRabbitMQ()
	defer rabbitChannel.Close()

	http.HandleFunc("/api/iot/car", iotCarHandle)
	http.HandleFunc("/api/iot/charge", iotChargeHandle)

	fmt.Println("🚀 Server listening on port 7070...")
	err := http.ListenAndServe(":7070", nil)
	if err != nil {
		log.Fatalf("❌ Server failed: %s", err)
	}
}
