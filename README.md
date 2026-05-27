# DIGIT eChallan Calculator (Go Re-implementation)

This is a production-ready, zero-deviation Go migration of the DIGIT `echallan-calculator` service. It is designed to act as a drop-in replacement in the DIGIT municipal-services ecosystem, preserving all API contracts, payloads, inter-service HTTP requests, and configurations.

## Architecture Mapping

| Java Component / Class | Go Package / File | Description |
|---|---|---|
| `org.egov.echallancalculation.web.controllers.ChallanCalController` | `web/controllers/challan_cal_controller.go` | Gin-based REST controller exposing `POST /echallan-calculator/v1/_calculate`. |
| `org.egov.echallancalculation.service.CalculationService` | `service/calculation_service.go` | Coordinates challan fee calculation, cancellation of existing bills, and demand generation. |
| `org.egov.echallancalculation.service.DemandService` | `service/demand_service.go` | Handles demand creation, update, and billing-service interactions. Replicates the custom threshold rounding logic perfectly. |
| `org.egov.echallancalculation.repository.DemandRepository` | `repository/repository.go` | Connects to Billing Service to create, update, and search demands. |
| `org.egov.echallancalculation.repository.ServiceRequestRepository` | `repository/repository.go` | Generic wrapper for executing external REST POST calls. |
| `org.egov.echallancalculation.util.CalculationUtils` | `util/util.go` | Construct search URLs and maps/fetches challan records from external services. |
| `org.egov.echallancalculation.config.ChallanConfiguration` | `config/config.go` | Handles application configuration variables (ports, endpoints, hosts). |
| Java model classes | `models/models.go` | Replicates request/response structures, preserving minor JSON spelling quirks (e.g. `CalulationCriteria`, `ResposneInfo`). |

## Getting Started

### Prerequisites
- Go 1.26.2 or higher
- Docker (optional, for containerization)

### Local Configuration
1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```
2. Adjust environment hostnames and ports in `.env` to match your local or cluster environment.

### Running Locally
Run the service directly using:
```bash
go run main.go
```
The service will start on port `8078` by default.

### Building
Compile the production-ready binary:
```bash
go build -o echallan-calculator main.go
```

### Docker deployment
Build the Docker image:
```bash
docker build -t egovio/echallan-calculator-go:1.0.0 .
```
Run the Docker container:
```bash
docker run -p 8078:8078 egovio/echallan-calculator-go:1.0.0
```

