# eChallan Ecosystem Runbook & Operational Guide

This document serves as the operational guide for deploying, testing, and managing the `echallan-services` and `echallan-calculator` microservices, completely updated for the Go migration.

## Architecture Overview
The system consists of two Go-based microservices replacing the legacy Spring Boot applications:
1. **eChallan Services**: The core data persistence layer and REST API (backed by PostgreSQL + GORM).
2. **eChallan Calculator**: The tax/penalty computation engine (stateless API orchestrator).

## 1. Environment Variables & Configuration
Both services enforce externalized configuration via `.env` files or Kubernetes ConfigMaps.

### eChallan Services
| Variable | Description | Default / Example |
|----------|-------------|-------------------|
| `SERVER_PORT` | HTTP port for the API | `8079` |
| `DB_HOST` | PostgreSQL Database Host | `localhost` |
| `DB_PORT` | PostgreSQL Database Port | `5432` |
| `DB_USER` | PostgreSQL User | `postgres` |
| `DB_PASSWORD` | PostgreSQL Password | `postgres` |
| `DB_NAME` | PostgreSQL Database Name | `echallan` |
| `KAFKA_BROKERS` | Kafka broker URL(s) | `localhost:9092` |
| `KAFKA_CHALLAN_CREATE_TOPIC` | Topic for created challans | `save-echallan` |
| `KAFKA_CHALLAN_UPDATE_TOPIC` | Topic for updated challans | `update-echallan` |
| `KEYCLOAK_URL` | Base URL for JWT Auth | `http://localhost:8080/auth/realms/digit` |

### eChallan Calculator
| Variable | Description | Default / Example |
|----------|-------------|-------------------|
| `SERVER_PORT` | HTTP port for the API | `8078` |
| `BILLING_HOST` | URL for DIGIT Billing Service | `http://localhost:8081` |
| `DEMAND_CREATE_ENDPOINT` | Path for Demand creation | `/billing-service/demand/_create` |
| `DEMAND_UPDATE_ENDPOINT` | Path for Demand update | `/billing-service/demand/_update` |
| `CHALLAN_HOST` | URL for eChallan Services | `http://localhost:8079` |
| `CHALLAN_SEARCH_ENDPOINT` | Path for Challan Search | `/api/v1/challans/search` |

## 2. Kafka Topics & Event Integration
The `echallan-services` acts as both a producer and consumer.
- **Producer:** When a challan is created/updated via the API, the service emits the full JSON payload to `save-echallan` or `update-echallan`.
- **Consumer:** A background Goroutine continually listens to `save-echallan`. Upon receiving an event, it triggers the MDMS notification service (e.g., sending an SMS or email). 

**At-Least-Once Delivery:** The Go consumers use manual offset commits via `CommitMessages()`. If the consumer panics or gracefully shuts down, uncommitted offsets are safely preserved.

## 3. Database Management (GORM)
The persistence layer uses GORM. 
- **Production Safety:** AutoMigrate is disabled. Ensure the following tables exist in the `echallan` database via Flyway or raw SQL migrations:
  - `eg_echallan`
  - `eg_challan_amount`
  - `eg_challan_address`
- **Pooling:** Connection pooling is configured (Max Open = 100, Max Idle = 10, Lifetime = 1h).

## 4. Authentication (JWT / OIDC)
The API routes are protected by a `KeycloakAuthMiddleware`.
- **Flow:** The client sends an `Authorization: Bearer <token>` header.
- **Validation:** The middleware contacts the `KEYCLOAK_URL` to validate the token signature and expiration.
- **Context Injection:** Upon success, the User ID and roles are injected into the Gin context for downstream RBAC.

## 5. Deployment Notes (Docker & Kubernetes)
### Docker Compose
To run the entire stack locally, including dependencies:
```bash
docker-compose up -d
```

### Kubernetes
Standard Deployments and Services should be applied via `kubectl`.
- Expose the services via an Ingress controller routing `/api/v1/challans` to `echallan-services` and `/api/v1/challans/calculate` to `echallan-calculator`.
- Store the `.env` variables in a `Secret` (for DB_PASSWORD) and a `ConfigMap` (for KAFKA_BROKERS).

## 6. Health Checks & Troubleshooting
### Health Probes
Liveness and Readiness probes should be configured to hit the root HTTP endpoint `/` (or a dedicated `/health` route). If the service responds with HTTP 200, it is healthy.

### Common Failures & Runbooks
**1. Connection Refused (PostgreSQL)**
- *Symptom:* Logs show `failed to connect to database: dial tcp [::1]:5432: connectex: No connection could be made`.
- *Fix:* Ensure `DB_HOST` is pointing to the correct Kubernetes service name or Docker network alias (e.g., `postgres` instead of `localhost`).

**2. Kafka Offset Desync**
- *Symptom:* Challans are created in the database, but notifications are not sending.
- *Fix:* Check the `echallan-services` logs for consumer panics. Use `kafka-consumer-groups` CLI to inspect the consumer group lag. If the consumer is stuck on a poison pill message, reset the offset manually.

**3. Calculator Timing Out**
- *Symptom:* `echallan-calculator` returns HTTP 500 when calculating.
- *Fix:* The calculator makes synchronous HTTP calls to `BILLING_HOST`. Ensure the Billing service is online. The calculator utilizes `context.Context` propagation, so if the billing service hangs, the request will safely timeout after 30 seconds rather than blocking indefinitely.
