# DIGIT eChallan Go Monorepo

This repository is the official Go-Lang migration of the DIGIT OSS eChallan ecosystem. It replaces the legacy Java Spring Boot services.

## Services Included

1. **echallan-calculator**: Stateless Go microservice responsible for generating dynamic demand formulas and computing tax heads. 
2. **echallan-services**: Core municipal microservice responsible for orchestrating Challan lifecycle events, communicating with PostgreSQL, and managing Kafka asynchronous event queues (DIGIT Persister Pattern).

## Architecture

Both services are built natively in Go, following the **Upgraded Go Template (UGT)** organizational layout. 
They each feature explicit dependency injection, distinct boundaries between transport (HTTP/Kafka) and business logic, and dedicated payload validation layers.

## Development

This repository utilizes **Go Workspaces (`go.work`)** to manage multiple modules simultaneously.
To develop locally:

1. Ensure Go 1.21+ is installed.
2. Open the root folder (`digit-echallan-go`) in your IDE.
3. The IDE will automatically read the `go.work` file and resolve dependencies for both `echallan-calculator` and `echallan-services`.

For specific service build instructions, environment variables, and Docker deployment steps, please read the `README.md` located inside each service folder.
