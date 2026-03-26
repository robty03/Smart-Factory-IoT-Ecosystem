package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"licenta-pubsub/internal/models"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
)

// --- METRICI PROMETHEUS ---
var (
	// Metrica pentru temperatură (folosită de alerta de Telegram)
	motorTemp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "motor_temperature_celsius",
		Help: "Temperatura actuală a motorului monitorizat",
	}, []string{"motor_id"})

	// Metrica pentru numărul total de mesaje procesate (bună pentru grafice)
	processedMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_processed_messages_total",
		Help: "Numărul total de mesaje salvate în baza de date",
	}, []string{"status"})
)

func connectDB() *sql.DB {
	for {
		db, err := sql.Open("postgres", os.Getenv("DB_URL"))
		if err != nil {
			log.Printf("⏳ Aștept DB: %v — reîncerc în 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		if err := db.Ping(); err != nil {
			log.Printf("⏳ DB nu răspunde: %v — reîncerc în 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("✅ Conectat la DB.")
		return db
	}
}

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

func ensureSchema(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS motor_telemetry (
			time        TIMESTAMPTZ       NOT NULL,
			motor_id    TEXT              NOT NULL,
			temperature DOUBLE PRECISION,
			vibration   DOUBLE PRECISION,
			current     DOUBLE PRECISION,
			rpm         DOUBLE PRECISION,
			noise_level DOUBLE PRECISION,
			status      TEXT
		);
		SELECT create_hypertable('motor_telemetry', 'time', if_not_exists => TRUE);
	`)
	if err != nil {
		log.Printf("⚠️ Avertisment schema DB: %v", err)
	}
}

func main() {
	// Pornire server de metrici pe portul 8081
	go func() {
		log.Println("📊 Server metrici pornit pe :8081/metrics")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("❌ Metrics server: %v", err)
		}
	}()

	db := connectDB()
	defer db.Close()

	ensureSchema(db)

	for {
		conn := connectRabbitMQ()

		ch, err := conn.Channel()
		if err != nil {
			log.Printf("❌ Eroare Channel: %v — reîncerc...", err)
			conn.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		q, err := ch.QueueDeclare("sensor_data", true, false, false, false, nil)
		if err != nil {
			log.Printf("❌ Eroare QueueDeclare: %v — reîncerc...", err)
			ch.Close()
			conn.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Printf("❌ Eroare Consume: %v — reîncerc...", err)
			ch.Close()
			conn.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		connClose := conn.NotifyClose(make(chan *amqp.Error, 1))

		log.Println("👷 Worker pornit... Aștept date de la motoare!")

		done := false
		for !done {
			select {
			case err := <-connClose:
				log.Printf("⚠️ Conexiune RabbitMQ pierdută: %v — reconectare...", err)
				done = true

			case d, ok := <-msgs:
				if !ok {
					log.Println("⚠️ Canal RabbitMQ închis — reconectare...")
					done = true
					break
				}

				var data models.MotorTelemetry
				if err := json.Unmarshal(d.Body, &data); err != nil {
					log.Printf("❌ Eroare parsare JSON: %v — mesaj respins", err)
					d.Nack(false, false)
					continue
				}

				// --- ACTUALIZARE METRICI ---
				// Trimitem temperatura către Prometheus (label-ul este ID-ul motorului)
				motorTemp.WithLabelValues(data.MotorID).Set(data.Temperature)

				// Inserare în TimescaleDB
				_, err := db.Exec(`
					INSERT INTO motor_telemetry 
						(time, motor_id, temperature, vibration, current, rpm, noise_level, status) 
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					data.Time, data.MotorID, data.Temperature, data.Vibration,
					data.Current, data.RPM, data.NoiseLevel, data.Status,
				)

				if err != nil {
					log.Printf("❌ Eroare DB: %v — mesaj requeue", err)
					processedMessages.WithLabelValues("error").Inc()
					d.Nack(false, true)
				} else {
					log.Printf("💾 SALVAT: %s (Temp: %.2f | Status: %s)", data.MotorID, data.Temperature, data.Status)
					processedMessages.WithLabelValues("success").Inc()
					d.Ack(false)
				}
			}
		}

		ch.Close()
		conn.Close()
		time.Sleep(3 * time.Second)
	}
}
