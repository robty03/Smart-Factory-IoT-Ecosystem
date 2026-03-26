# 🏭 Smart Factory IoT Ecosystem: Monitoring and Distributed Processing in Kubernetes

This project represents a comprehensive **End-to-End IoT** solution developed for a Bachelor's Thesis. The system simulates, collects, processes, and monitors real-time telemetry from industrial machinery (CNC motors, pumps, conveyor belts) within a "Smart Factory" environment.

## 🏗️ System Architecture

The system is built on a microservices architecture, fully orchestrated in **Kubernetes**, ensuring high availability, scalability, and self-healing capabilities.

### Data Pipeline Flow:
1. **Source (Simulator)**: An independent utility that generates synthetic telemetry and simulates failure scenarios (overheating, mechanical wear).
2. **Ingestion (Producer)**: A Go-based HTTP Gateway that receives JSON payloads and publishes them to RabbitMQ.
3. **Broker (RabbitMQ)**: Decouples ingestion from processing, ensuring message persistence and reliable delivery.
4. **Processing (Worker)**: Consumes data, performs validation, and persists records into the database.
5. **Storage (TimescaleDB)**: A PostgreSQL-based SQL database optimized specifically for time-series data.
6. **Observability (LGTM Stack)**: Monitoring via Prometheus, logging via Loki/Promtail, and visualization through Grafana dashboards.
7. **Alerting (Telegram)**: Instant notifications for critical incidents sent via Alertmanager.

---

## 📂 Project Structure and File Roles

### 📂 `cmd/` (Business Logic)
* **`producer/`**: Gateway service. Exposes the `/api/sensor` endpoint and internal Prometheus metrics (`/metrics`).
* **`worker/`**: Data processor. Responsible for writing to TimescaleDB and managing asynchronous tasks from RabbitMQ.
* **`simulator/`**: Testing utility (formerly CLI). Generates data based on physical models (Normal, Wear, Thermal Hazard).

### 📂 `k8s/` (Infrastructure as Code)
* **`00-base.yaml`**: Namespace definition and basic resource allocation.
* **`apps/`**: Kubernetes manifests for Producer and Worker, defining replicas and restart policies.
* **`infrastructure/`**: Setup for support services: RabbitMQ (broker) and TimescaleDB (database).
* **`monitoring/`**: The Observability stack.
    * `03-alertmanager.yaml`: Telegram integration, unique Incident ID (Fingerprint) management, and timezone logic (UTC vs RO Time).
    * `04-prometheus-rules.yaml`: Mathematical rules defining alert thresholds.

### 📂 `internal/`
* **`internal/models/`**: The single source of truth for data structures (`MotorTelemetry`), ensuring consistency across all components.

---

## 🚀 Installation and Operations (via Makefile)

The project includes a **Makefile** to automate all lifecycle processes:

| Command | Description |
| :--- | :--- |
| `make deploy` | Deploys the entire ecosystem (Databases, Broker, Apps, Monitoring) to the K8s cluster. |
| `make simulate` | Starts the data injector with built-in anomaly scenarios. |
| `make status` | Checks the real-time health and status of all cluster pods. |
| `make undeploy` | Completely removes all allocated resources from the Kubernetes cluster. |

---

## 🔔 Observability and Incident Response

The system places a heavy emphasis on proactive monitoring:
* **Telegram Notifications**: Alerts include the incident ID (original Prometheus Fingerprint) for academic traceability.
* **Time Management**: Reporting is done in the **UTC** standard, with an informative note for **Romania Time (UTC+2h)**, eliminating confusion during log audits.
* **Log Analytics**: Loki/Promtail integration allows for direct correlation between alerts and application logs within Grafana.

---

## 🛠️ Technology Stack
* **Language**: Go (Golang)
* **Orchestration**: Kubernetes (K8s)
* **Messaging**: RabbitMQ (AMQP)
* **Database**: TimescaleDB (PostgreSQL)
* **Monitoring**: Prometheus, Grafana, Loki, Alertmanager
* **Communication**: Telegram Bot API, REST HTTP

---

### ✅ Roadmap (Advanced Components):
1. **HPA (Horizontal Pod Autoscaler)**: Automatic scaling of workers based on processing load.
2. **K8s Secrets**: Secure management of Telegram tokens and database credentials.
3. **Stress Testing**: Evaluation of system performance under massive data throughput using the simulator.
