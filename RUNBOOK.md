# eChallan System Runbook

This document serves as the operational guide for deploying, testing, and managing the `echallan-services` and `echallan-calculator` microservices.

## Architecture Overview
The system consists of two Go-based microservices:
1. **eChallan Services**: The core data persistence layer and REST API (backed by PostgreSQL + GORM).
2. **eChallan Calculator**: The tax/penalty computation engine (stateless).

Both services use:
- **Keycloak** for JWT/OIDC authentication.
- **Kafka** for asynchronous messaging (receipts, notifications).

## 1. Local Development
To run the entire stack locally, including dependencies:
```bash
docker-compose up -d
```
This spins up:
- PostgreSQL (Port: 5432)
- Zookeeper (Port: 2181)
- Kafka (Port: 9092)
- eChallan Services (Port: 8079)
- eChallan Calculator (Port: 8078)

## 2. Keycloak Configuration
The services expect a running Keycloak server. By default, they look for:
- **URL**: `http://localhost:8080`
- **Realm**: `digit`
- **Clients**: `echallan-services`, `echallan-calculator`

**Required Setup:**
1. Create the `digit` realm in Keycloak.
2. Ensure the JWT tokens include a `sub` (User ID) and `realm_access.roles` (for RBAC).

## 3. Database Management (GORM)
The persistence layer uses GORM. Currently, AutoMigrate is disabled for production safety. 
Ensure the following tables exist in the `echallan` database:
- `eg_echallan`
- `eg_challan_amount`
- `eg_challan_address`

## 4. Running Unit Tests
Unit tests use the standard Go `testing` package along with `testify` for mocks and assertions.
```bash
# Run tests for core services
cd echallan-services
go test ./internal/... -v

# Run tests for calculator
cd echallan-calculator
go test ./internal/... -v
```

## 5. Kubernetes Deployment
Basic manifests are provided in the `deployments/k8s/` directory of each service.
```bash
kubectl apply -f echallan-services/deployments/k8s/
kubectl apply -f echallan-calculator/deployments/k8s/
```
Make sure to create a ConfigMap or Secret for the environment variables (`DB_PASSWORD`, `KEYCLOAK_URL`, etc.) before deploying to a cluster.
