package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"licenta-pubsub/internal/models"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
)

func connectRabbitMQ() *amqp.Connection {
	for {
		conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
		if err != nil {
			log.Printf("⏳ Aștept RabbitMQ: %v — reîncerc în 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("✅ Conectat la RabbitMQ.")
		return conn
	}
}

func main() {
	conn := connectRabbitMQ()
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Eroare Channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("sensor_data", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Eroare QueueDeclare: %v", err)
	}

	// Detectăm pierderea conexiunii și oprim graceful
	connClose := conn.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		if err := <-connClose; err != nil {
			log.Fatalf("⚠️ Conexiune RabbitMQ pierdută: %v — restarting container...", err)
		}
	}()

	http.HandleFunc("/api/sensor", func(w http.ResponseWriter, r *http.Request) {
		var data models.MotorTelemetry
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if data.Time.IsZero() {
			data.Time = time.Now()
		}

		body, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "marshal error", http.StatusInternalServerError)
			return
		}

		err = ch.PublishWithContext(
			context.Background(), "", q.Name, false, false,
			amqp.Publishing{ContentType: "application/json", Body: body},
		)
		if err != nil {
			log.Printf("❌ Eroare publish: %v", err)
			http.Error(w, "publish error", http.StatusInternalServerError)
			return
		}

		log.Printf("📥 PRIMIT: %s | Temp: %.1f | Status: %s", data.MotorID, data.Temperature, data.Status)
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Println("🚀 Producer activ pe :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ HTTP server: %v", err)
	}
}
