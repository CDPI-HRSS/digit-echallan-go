# DIGIT e-Challan: Go Conversion Project 

## Project Overview
This repository contains the Go (Golang) conversion of the **DIGIT e-Challan** microservices. Originally written in Java (Spring Boot), this project acts as a "drop-in replacement" for the `echallan-services` and `echallan-calculator` modules within the eGov DIGIT ecosystem.

This is a minor project developed by students at **KLE Tech** in collaboration with the DPI (Digital Public Infrastructure) initiative.

##  Objectives
* **Performance Optimization:** Drastically reduce the memory footprint and container startup times by replacing the Java Virtual Machine (JVM) with compiled Go binaries.
* **Strict API Contract Adherence:** Maintain 100% compatibility with the existing DIGIT OpenAPI/Swagger contracts (e.g., `RequestInfo`, `ResponseInfo`).
* **Seamless Integration:** Ensure the new Go services connect flawlessly to the existing core infrastructure (Zuul API Gateway, MDMS, User Service, and PostgreSQL).

## Project Roadmap
-  **Phase 1: Baseline Architecture Deployment** - Provisioned Azure Virtual Machine.
  - Deployed stateful infrastructure (Postgres, Redis, Kafka).
  - Deployed Java baseline of Core and Business services via Kubernetes/Helm.
  - Captured Postman baseline metrics for API routing and OAuth2 security.
-  **Phase 2: The Go Rewrite (In Progress)**
  - Map Java Swagger contracts to Go structs.
  - Rewrite `echallan-calculator` penalty logic in Go.
  - Rewrite `echallan-services` core logic in Go.
-  **Phase 3: Containerization & Swap**
  - Dockerize Go applications.
  - Perform live Kubernetes pod swap (Java -> Go).
  - Final end-to-end integration testing.
