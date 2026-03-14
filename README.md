# GoFlow: High-Performance Event Tracking System

GoFlow is a distributed, high-load event tracking system built with **Go**, designed to ingest, process, and store massive amounts of event data in real-time.

## 🚀 Architecture Overview
The system follows a modern microservices pattern to ensure scalability and fault tolerance:
* **Gateway Service**: High-speed API entry point that validates and pushes incoming events to Kafka.
* **Apache Kafka**: Serves as the robust message backbone, decoupling ingestion from processing.
* **Worker Service**: Consumer group that processes events from Kafka and performs bulk inserts into storage.
* **ClickHouse**: OLAP database used for high-speed analytical storage and querying.
* **Monitoring**: Integrated **Prometheus** and **Grafana** stack for real-time visibility into system metrics.

## 🛠 Tech Stack
- **Language**: Go (Golang)
- **Message Broker**: Apache Kafka
- **Storage**: ClickHouse
- **Orchestration**: Kubernetes (K8s)
- **Load Testing**: k6
- **Monitoring**: Prometheus & Grafana

## 📦 Infrastructure & Deployment
The project includes a complete set of Kubernetes manifests for local or cloud deployment:
- **`01-clickhouse.yaml`**: ClickHouse server setup.
- **`02-kafka.yaml`**: Kafka broker and controller configuration.
- **`03-apps.yaml`**: Deployment specs for Gateway and Worker services.
- **`k6-job.yaml`**: Automated stress-testing configuration.

## 📈 Performance Testing
The system is designed to handle high RPS (Requests Per Second). Stress tests are performed using **k6**, simulating thousands of virtual users to validate the stability of the Kafka ingestion pipeline and ClickHouse write speeds.
