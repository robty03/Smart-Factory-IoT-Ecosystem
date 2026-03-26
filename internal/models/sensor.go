package models

import "time"

type MotorTelemetry struct {
	Time        time.Time `json:"time"`
	MotorID     string    `json:"motor_id"`
	Temperature float64   `json:"temperature"`
	Vibration   float64   `json:"vibration"`
	Current     float64   `json:"current"`
	RPM         float64   `json:"rpm"`
	NoiseLevel  float64   `json:"noise_level"`
	Status      string    `json:"status"`
}
