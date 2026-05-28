# DIGIT eChallan Ecosystem (Go)

This monorepo contains the modern, high-performance Go microservices for the DIGIT eChallan ecosystem. It replaces the legacy Java Spring Boot services with an idiomatic, Upgraded Go Template (UGT) compliant architecture.

## Microservices

| Service | Description | Port |
|---------|-------------|------|
| [echallan-services](./echallan-services) | Core module for Challan lifecycle, search, and asynchronous state transitions via Kafka. | `8079` |
| [echallan-calculator](./echallan-calculator) | Mathematical computation engine for tax estimation, penalty calculations, and billing integration. | `8078` |

## Architecture Highlights
- **Strict UGT Compliance**: Enforces explicit separation of domain, repository, service, and transport layers.
- **High Performance**: Built on `gin-gonic` and `sqlx`.
- **Event-Driven Resilience**: Utilizes `segmentio/kafka-go` with semaphore-bounded connection pooling to replicate the eGov Persister pattern.
- **Security-First**: Fully immune to SQL Injection and protected against OOM (Out of Memory) payloads.

## Local Development (Orchestration)

You can spin up the entire DIGIT eChallan ecosystem, along with its backing data stores (Postgres & Kafka), using Docker Compose.

```bash
docker-compose up --build
```

### Individual Service Setup
To build or test services individually without Docker, navigate to their respective directories and consult their inner `README.md`.

```bash
cd echallan-services
go run cmd/echallan-services/main.go
```
