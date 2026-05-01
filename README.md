# DIGIT e-Challan: Go Conversion Project 

## Project Overview
This repository contains the Go (Golang) conversion of the **DIGIT e-Challan** microservices. Originally written in Java (Spring Boot), this project acts as a "drop-in replacement" for the `echallan-services` and `echallan-calculator` modules within the eGov DIGIT ecosystem.

This is a minor project developed by students at **KLE Tech** in collaboration with the DPI (Digital Public Infrastructure) initiative.

##  Objectives
* **Performance Optimization:** Drastically reduce the memory footprint and container startup times by replacing the Java Virtual Machine (JVM) with compiled Go binaries.
* **Strict API Contract Adherence:** Maintain 100% compatibility with the existing DIGIT OpenAPI/Swagger contracts (e.g., `RequestInfo`, `ResponseInfo`).
* **Seamless Integration:** Ensure the new Go services connect flawlessly to the existing core infrastructure (Zuul API Gateway, MDMS, User Service, and PostgreSQL).

## Project Roadmap
**Phase 1: Baseline Architecture Deployment** - Provisioned Azure Virtual Machine.
  - Deployed stateful infrastructure (Postgres, Redis, Kafka).
  - Deployed Java baseline of Core and Business services via Kubernetes/Helm.
  - Captured Postman baseline metrics for API routing and OAuth2 security.
**Phase 2: The Go Rewrite (In Progress)**
  - Map Java Swagger contracts to Go structs.
  - Rewrite `echallan-calculator` penalty logic in Go.
  - Rewrite `echallan-services` core logic in Go.
**Phase 3: Containerization & Swap**
  - Dockerize Go applications.
  - Perform live Kubernetes pod swap (Java -> Go).
  - Final end-to-end integration testing.

  🛠️ Tech Stack
Language: Go (Golang)

Framework: Gin / Fiber (TBD)

Infrastructure: Kubernetes, Helm, Docker, Azure

Databases: PostgreSQL, Redis

Team
Raghavendra Rao Kulkarni-  DevOps & Cloud Infrastructure

Sneha Channappagoudar - Core Services & Database Integration

Harshitha M B - Business Logic & API Translation

Spurti SMK - API Testing, Environment setup


---

### **3. How to initialize and push this to GitHub (Terminal)**

Once you have created the empty repository on GitHub (let's assume you named it `digit-echallan-go`), open your terminal on your local laptop, create a folder, and run these commands to push your first commit:

```bash
# 1. Create your project folder and enter it
mkdir digit-echallan-go
cd digit-echallan-go

# 2. Initialize Git
git init

# 3. Create the README file (Paste the markdown above into this file)
nano README.md  

# 4. Add the file to version control
git add README.md

# 5. Commit the file
git commit -m "Initial commit: Added professional README and project roadmap"

# 6. Link it to your GitHub repository (Replace YOUR_USERNAME with your actual GitHub username)
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/digit-echallan-go.git

# 7. Push it to the cloud!
git push -u origin main
