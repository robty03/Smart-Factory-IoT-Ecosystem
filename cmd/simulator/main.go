package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"licenta-pubsub/internal/models"
)

// httpClient cu timeout explicit — evităm blocarea pe http.Post default
var httpClient = &http.Client{Timeout: 3 * time.Second}

func evalStatus(temp, vib float64) string {
	if temp > 80.0 || vib > 15.0 {
		return "CRITICAL"
	} else if temp > 60.0 || vib > 5.0 {
		return "WARNING"
	}
	return "OK"
}

func main() {
	apiURL := "http://localhost:8080/api/sensor"
	log.Println("🛠️ Simulator Smart Factory pornit! Se generează date...")

	anomalyTemp := 50.0
	anomalyVib := 2.0

	for {
		// Scenariul 1: Normal
		temp1 := 45.0 + rand.Float64()*5
		vib1 := 2.0 + rand.Float64()*1
		sendData(apiURL, models.MotorTelemetry{
			MotorID:     "MOTOR_CNC_01",
			Temperature: temp1,
			Vibration:   vib1,
			Current:     12.0 + rand.Float64()*2,
			RPM:         1500.0 + rand.Float64()*10,
			NoiseLevel:  60.0 + rand.Float64()*5,
			Time:        time.Now(),
			Status:      evalStatus(temp1, vib1),
		})

		// Scenariul 2: Uzură progresivă
		anomalyVib += 0.2
		if anomalyVib > 15.0 {
			anomalyVib = 2.0
		}
		temp2 := 50.0 + rand.Float64()*5
		vib2 := anomalyVib + rand.Float64()*2
		sendData(apiURL, models.MotorTelemetry{
			MotorID:     "MOTOR_BANDA_02",
			Temperature: temp2,
			Vibration:   vib2,
			Current:     13.0 + rand.Float64()*2,
			RPM:         1480.0 - anomalyVib,
			NoiseLevel:  70.0 + (anomalyVib * 2),
			Time:        time.Now(),
			Status:      evalStatus(temp2, vib2),
		})

		// Scenariul 3: Pericol termic
		anomalyTemp += 1.5
		if anomalyTemp > 95.0 {
			anomalyTemp = 50.0
		}
		temp3 := anomalyTemp + rand.Float64()*3
		vib3 := 3.0 + rand.Float64()*1
		sendData(apiURL, models.MotorTelemetry{
			MotorID:     "MOTOR_POMPA_03",
			Temperature: temp3,
			Vibration:   vib3,
			Current:     15.0 + (anomalyTemp / 10),
			RPM:         1500.0,
			NoiseLevel:  65.0,
			Time:        time.Now(),
			Status:      evalStatus(temp3, vib3),
		})

		time.Sleep(1 * time.Second)
	}
}

func sendData(url string, data models.MotorTelemetry) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("⚠️ Eroare marshal: %v", err)
		return
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("⚠️ Eroare trimitere [%s]: %v", data.MotorID, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("✅ Trimis: [%s] %s | Temp: %.1f°C | Vib: %.1fmm/s | Curr: %.2fA | RPM: %.2f | Noise: %.2fdB",
		data.Status, data.MotorID, data.Temperature, data.Vibration, data.Current, data.RPM, data.NoiseLevel)
}
