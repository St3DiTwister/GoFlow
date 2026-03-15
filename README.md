# GoFlow: High-Performance Event Tracking System

GoFlow is a distributed, high-load event tracking system built with **Go**, designed to ingest, process, and store massive amounts of event data in real-time.

## 🚀 Performance Achievement
> **Benchmark Result:** The system successfully handles **71,000+ RPS** (Requests Per Second) on a single node with **0% data loss**.

* **Ingestion Rate:** ~71k msg/s (Gateway → Kafka)
* **Processing Rate:** ~15k-20k msg/s (Worker → ClickHouse)
* **Resiliency:** 100% data integrity during extreme traffic spikes. The architecture utilizes Kafka as a "shock absorber," allowing the system to buffer millions of events during peak loads.

## 🏗 Architecture & Design
The system is built on an asynchronous processing pipeline to handle backpressure effectively:
* **Gateway Service**: Ultra-fast entry point in Go. It prioritizes ingestion speed, acknowledging requests with `202 Accepted` in ~10ms.
* **Apache Kafka**: Serves as the robust message backbone (configured with 10 partitions). It decouples ingestion from storage, preventing database overload.
* **Worker Service**: Consumer group with **batch processing** logic. It performs optimized bulk inserts (up to 7,000 rows) into ClickHouse to minimize CPU and I/O overhead.
* **ClickHouse**: OLAP database providing lightning-fast analytical storage for millions of events.

## 🛠 Tech Stack
- **Language**: Go (Golang)
- **Message Broker**: Apache Kafka
- **Storage**: ClickHouse
- **Orchestration**: Kubernetes (K8s) & Docker Compose
- **Monitoring**: Prometheus & Grafana
- **Load Testing**: k6

## 📊 Monitoring & Observability
The integrated observability stack provides real-time visibility into:
- **Throughput**: Inbound RPS vs. Processing Rate.
- **Kafka Lag**: Real-time tracking of buffered messages during bursts.
- **System Health**: Resource utilization metrics for all components.

## 📦 Infrastructure
The project includes production-ready manifests for both **Docker Compose** and **Kubernetes (K8s)**, ensuring flexible deployment across different environments.