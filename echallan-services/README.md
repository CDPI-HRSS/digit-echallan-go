# eChallan Services Go Migration

## Overview
This service is the Go language migration of the DIGIT `echallan-services`. It conforms exactly to the DIGIT OSS Migration Handbook, utilizing a strict layered architecture, explicit routing via Gin, and asynchronous persistence via Kafka (The DIGIT Persister Pattern).

## Prerequisites
- Go 1.21+
- PostgreSQL
- Kafka (for local testing of async flows)

## Environment Variables
Ensure the following variables are set in `configs/.env`:
- `SERVER_PORT` (default: 8082)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `KAFKA_BROKERS`
- `PERSISTER_SAVE_CHALLAN_TOPIC`

## Run Locally
1. `go mod tidy`
2. `go run cmd/echallan-services/main.go`

## Deployment
Build the docker container:
`docker build -t echallan-services-go:v1 -f deployments/Dockerfile .`
