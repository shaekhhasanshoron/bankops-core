# BankOps-Core: System Architecture & Design Documentation

## Index
* [Introduction](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#1-introduction)
* [Business Architecture & Service Design](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#2-business-architecture--service-design)
   * [Core Business Domain & User Roles](https://github.com/shaekhhasanshoron/bankops-core/tree/master?tab=readme-ov-file#21-core-business-domain--user-roles)
   * [Service Description](https://github.com/shaekhhasanshoron/bankops-core/tree/master?tab=readme-ov-file#22-service-description)
   * [Service Responsibilities](https://github.com/shaekhhasanshoron/bankops-core/tree/master?tab=readme-ov-file#23-service-responsibilities)
   * [Why This Separation?](https://github.com/shaekhhasanshoron/bankops-core/tree/master?tab=readme-ov-file#24-why-this-separation)
   * [Inter-Service Communication](https://github.com/shaekhhasanshoron/bankops-core/tree/master?tab=readme-ov-file#25-inter-service-communication)
* [Technical Overview](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#3-technical-overview)
  * [Architectural Patterns & Principles](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#31-architectural-patterns--principles)
  * [Production Excellence and Key Features](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#32-production-excellence-and-key-features)
    * [Resilience & Reliability](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#321-resilience--reliability)
    * [Data Integrity & Security](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#322-data-integrity--security)
    * [Observability](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#323-observability)
    * [Technology and Tools](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#324-technology-and-tools)
* [Project Structure & Maintainability](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#4-project-structure--maintainability)
  * [Why This Structure?](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#41-why-this-structure)
* [Testing, Security & Observability](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#5-testing-security--observability)
  * [Testing Strategy and Coverage](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#51-testing-strategy-and-coverage)
  * [Security](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#52-security)
  * [Observability](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#53-observability)
* [Installation & Deployment](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#6-installation--deployment)
  * [Install locally](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#61-install-locally)
  * [Run on Docker](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#62-run-on-docker)
  * [Kubernetes](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#63-kubernetes)
* [Live Demo](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#7-live-demo)
* [API Documentation](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#8-api-documentation)
* [Future Enhancements](https://github.com/shaekhhasanshoron/bankops-core?tab=readme-ov-file#9-future-enhancements)

## 1. Introduction
This document provides a detailed overview of the architecture and implementation of BankOps-Core, 
a microservice-based backend API designed to serve core banking operations for bank employees. The objective was to 
build a system that is not only functionally reliable but also resilient, scalable, secure, and capable of enduring over time. 

The idea was to create a robust foundation that could upgrade and adapt to the evolving needs of the bank, 
while avoiding vendor lock-in and ensuring that each component remains independently scalable and maintainable.

## 2. Business Architecture & Service Design

The system is divided into four distinct services, each representing a clearly defined domain. 
This separation allows each team to own their domain logic independently, enabling parallel development and reducing operational complexity.

### 2.1 Core Business Domain & User Roles
The system is designed for internal bank employees as the primary users, enabling them to manage core banking 
operations through a secure, role-based interface.

**Authentication & Security:**
* All API access requires a Bearer token obtained through the login endpoint
* Tokens have a default lifespan of **15 minutes** for security
* Role-based access control (RBAC) is enforced across all operations


**User Roles & Privileges:**
* **Admin Employees:** Possess full system privileges. Can manage other employees, customers, bank accounts, and financial transactions.
* **Editor Employees:** Have all privileges except employee management. Can create, update, and delete customers, accounts, and process transactions.
* **Viewer Employees:** Have read-only access. Can view customer information, account details, and transaction history but cannot make any modifications.

**Core Banking Operations:**
* Customer management (create, update, view)
* Bank account management (create multiple accounts per customer)
* Financial transactions (transfer between accounts, withdrawals, deposits)
* Comprehensive audit trails for all operations

### 2.2 Service Description

Here's, the top level overview of the system:

![](https://github.com/shaekhhasanshoron/bankops-core/blob/master/static/system-overview.jpeg)

**Gateway Service**: This service serves as the central access point for all employee requests. 
It acts as a proxy, routing requests to the respective internal services, ensuring security, role-based access control (RBAC), and observability.

**Auth Service**: The authentication service manages employee identities, roles, and permissions. 
It provides authentication for all employees and ensures that access is granted based on predefined roles.

**Account Service**: This service is dedicated to managing customers and their bank accounts. 
It ensures that all account-related data (such as balances and account details) are securely 
stored and easily accessible by authorized employees.

**Transaction Service**: This is the core service that is responsible for managing transactions between accounts. 
It facilitates deposits, withdrawals, and transfers. 
Additionally, it tracks transaction history and ensures that all transactions are handled correctly.

### 2.3 Service Responsibilities

**Gateway Service:**
* Acts as a single entry point for all employee requests.
* Ensuring internal services can focus purely on business logic.
* Routes requests to appropriate internal services (Auth and Account services).
* Provides logging, and observability features for monitoring.
* It authorizes the JWT token to reduce load on the Auth service.

**Auth Service:**
* Ensures authentication and authorization via JWT tokens.
* Handles employee creation, authentication, and role management.
* Admins can create, update, or delete employee accounts.
* Supports RBAC (Role-Based Access Control) where employees can have different roles (`admin`, `editor`, `viewer`).
* Issues JWT tokens upon successful login to authenticate employees for future requests.

**Account Service:**
* Manages customer data and account-related operations.
* Handles account creation, viewing of account details.
* Tracks all events like customer creation for auditing. 

**Transaction Service:**
* Manages financial transactions between accounts.
* Facilitates deposits, withdrawals, and transfers between accounts.
* Ensures that multiple concurrent transactions for the same account are handled correctly.
* Fetches transaction history based on parameters like customer ID, account number, and date range.
* Handles transaction recovery and consistency during server failures.

### 2.4 Why This Separation?
The separation of concerns into distinct services allows for better scalability, maintainability, and flexibility. 

Each service has a clear responsibility:
* The **Gateway Service** provides a unified interface to the outside world while keeping internal logic separate.
* The **Auth Service** encapsulates authentication and authorization, ensuring role-based access without affecting other services.
* The **Account Service** is focused solely on account management, making it more modular and decoupled from transaction logic.
* The **Transaction Service** is independently responsible for transaction management, 
  providing a dedicated space for handling financial transactions, with added flexibility for enhancements 
  such as transaction recovery and concurrency handling.

### 2.5 Inter-Service Communication
* **External (Employee to System)**: All access is through the Gateway via RESTful HTTP APIs. 
This provides a familiar and well-structured interface for clients.
* **Internal (Service to Service)**: The Gateway communicates with all backend services using gRPC. 
Crucially, the Transaction Service acts as an orchestrator, making gRPC calls to the Account Service to validate accounts 
and update balances during a transaction.
**gRPC** is integrated because of its performance, strong contract-first API definitions via Protocol Buffers, and built-in connection pooling.
* **Asynchronous Events**: For decoupled, event-driven workflows, **message queue** (e.g kafka) has been integrated.
All the service publishes events (e.g., `TransactionCompleted`, `CustomerCreated`, `EmployeeCreated`) to dedicated topics. 
This allows future services (like a Notification Service or Analytics Service) to react to these events without the event produce services 
needing to know they exist, solving the problem of tight coupling and enabling system-wide scalability.

## 3. Technical Overview

The focus was on creating a system that is not just a prototype but a production-ready platform. 
Every pattern and technology was chosen to solve a specific production challenge.

Here's the system architecture of **"Account Service"**:

![](https://github.com/shaekhhasanshoron/bankops-core/blob/master/static/system-architecture.jpeg)

### 3.1 Architectural Patterns & Principles
**Microservice Foundation with Bounded Contexts**: 

The services are modeled around clear business domains (Auth, Account, Transaction). This allows each service to have its own 
independent data model, technology, and team.

**Hexagonal Architecture (Ports & Adapters):** 

This is the cornerstone of the design to prevent vendor lock-in. The core business logic is completely isolated 
from external concerns like databases and APIs. It only interacts with abstract interfaces (Ports). 
The concrete implementations (Adapters) for GORM, gRPC, or Kafka can be swapped without touching the core domain. 
This makes the system incredibly adaptable and testable.

**Domain-Driven Design (DDD):** 

Banking has a complex, rule-heavy domain. By modeling entities like Customer, Account, and Transaction as rich objects 
with their own business rules and invariants, the code becomes a direct reflection of the business. 
This makes it easier to maintain and ensures that complex rules, like those governing transactions, are encapsulated correctly.

**Saga Orchestrator Pattern:**

A fund transfer is a perfect example of a long-running business process (a Saga) that spans multiple services. 
The Transaction Service acts as the orchestrator, executing the transfer in a series of steps (debit, credit) via gRPC. 
If any step fails (e.g., insufficient funds), the orchestrator triggers compensating transactions (e.g., a reverse debit) 
to roll back the operation. This elegantly solves the problem of maintaining data integrity in a distributed system 
without resorting to complex, tightly-coupled distributed transactions.

**Repository Pattern:** 

This pattern is used to abstract the data access layer. The business logic calls methods on a `CustomerRepo` interface, not a database. 
This cleanly separates the "what" (get a customer by ID) from the "how" (using GORM to query PostgreSQL/SQlite/MongoDB). 
It makes the code testable and easy to mock the repository that allows changing the underlying database with minimal impact.

**Singleton Pattern:** 

This is applied for infrastructure components that must have a single instance. 
The configuration manager, the database connection pool, the Kafka producer, and the logging instance are all initialized as singletons. 
This solves the problem of wasteful resource duplication and ensures consistent, coordinated access across the application.

**API Gateway Pattern:** 

This creates one strong entry point for the system. It centralizes critical cross-cutting concerns like security and observability, 
ensuring they are applied consistently without every internal service having to re-implement them.

### 3.2 Production Excellence and Key Features

#### 3.2.1 Resilience & Reliability

* **State Monitoring & Health Checks:** Each service exposes health endpoints, allowing Kubernetes or a load balancer to 
monitor its liveness and readiness, ensuring traffic is only sent to healthy instances.

* **Automatic Transaction Recovery:** Background jobs are created that scans for and recovers transactions stuck in an intermediate state 
due to a service crash. This ensures data consistency and solves the problem of manual intervention after a failure.

* **Resilient Messaging:** Kafka health monitor with exponential backoff reconnection 
ensures self-healing from network partitions or broker downtime.

* **Exponential Backoff & Retry:** For gRPC calls and kafka health check, retry mechanisms is implemented with exponential backoff. 
This makes the system resilient to temporary network glitches or brief downtime of a dependent service.

* **Graceful Shutdown:** Upon receiving a shutdown signal, the service stops accepting new requests, 
finishes processing current ones, and cleanly closes all database and network connections. 
This prevents data corruption and dropped transactions during deployments.

#### 3.2.2 Data Integrity & Security

* **ACID Transactions:** All financial operations, especially fund transfers, are wrapped in database transactions. 
This guarantees that operations are Atomic, Consistent, Isolated, and Durable.

* **Optimistic Locking:** To handle concurrent requests for the same account, **version** field were used. 
If two requests try to update the same account version, the latter one fails and must retry. 
This prevents race conditions and lost updates without the performance penalty of full database locks.

* **Multi-level Locking:** For huge requests at the same time for a same account, optimistic locking has been implemented 
with explicit database locks to ensure absolute consistency.

* **JWT Authentication & RBAC:** The Gateway validates JWT tokens and enforces role-based access. 
An employee with a "`viewer`" role cannot perform actions reserved for an "`editor`" securing the system from unauthorized use.

* **SQL Injection Prevention:** The GORM ORM and prepared statements automatically sanitize all inputs, 
making SQL injection attacks impossible.

#### 3.2.3 Observability

* **Prometheus:** Prometheus is integrated to collect and expose business and system metrics (e.g., number of transactions, active connections). 
This allows for real-time monitoring and alerting.

* **OpenTelemetry (OTel):** Distributed tracing added to track a request as it flows through the Gateway, Auth, and Account services. 
This is invaluable for debugging complex issues and understanding performance bottlenecks across service boundaries.

* **Structured Logging (Zerolog):** All logs are structured as JSON, making them easy to parse, search, and analyze 
in tools like Elasticsearch or Loki. Errors are categorized for quick filtering.

#### 3.2.4 Technology and Tools

* **Language & Frameworks:** 
  * [Go](https://go.dev/) for its performance, excellent concurrency model, and robust standard library. 
  * [Gin](https://gin-gonic.com/) powers the HTTP server in the Gateway. 
  * [Testify](https://github.com/stretchr/testify) is used for a clean and powerful testing experience.

* **Event Streaming:** 
  * [Kafka Go](https://github.com/segmentio/kafka-go) is used to reliably publish domain events from the Account service.

* **Configuration Management**: 
  * [Koanf](https://github.com/knadh/koanf) allows for flexible configuration from multiple sources (env vars, files (injected through secret management tools like HashiCorp Vault)), 
  making the application easy to deploy in any environment.

* **Development Experience:** 
  * [Air](https://github.com/air-verse/air) for live reloading during development 
  * [golangci-lint](https://golangci-lint.run/) to enforce code quality and consistency automatically.

## 4. Project Structure & Maintainability

The project structure is a direct physical manifestation of the Hexagonal Architecture, designed for clarity and ease of development.

````
transaction-service/
├── .air.toml                                    # Live reload config for development
├── go.mod                                       # Go modules for dependencies
├── Dockerfile                                   # Dockerfile for containerization
├── .env                                         # Environment variables for development
├── cmd/
│   └── txsvc/
│       └── main.go                              # Entry point to the service
├── api/
│   ├── proto/                                   # gRPC service definitions
│   └── protogen/                                # gRPC generated files
├── internal/
│   ├── domain/                                  # Core domain logic (Transaction...)
│   ├── app/                                     # Business logic and use cases and unit tests
│   ├── ports/                                   # Interfaces for external communication (repos, message publishers, grpc clients)
│   ├── adapters/                                # Infrastructure components (e.g., repositories, message publisher, gRPC client implementation)
│   │   ├── repo/                                # Database interactions
│   │   ├── grpc/                                # gRPC client adapater
│   │   └── message_publisher/                   # Event/message publishing logic
│   ├── grpc/                                    # gRPC server and handlers
│   │   ├── server.go                            # gRPC server setup
│   │   └── interceptors.go                      # gRPC interceptors (metrics, tracing...)
│   ├── http/                                    # HTTP server, health check, metrics export
│   ├── observability/                           # Metrics and tracing (Prometheus, OpenTelemetry)
│   │   ├── metrics.go                           # Prometheus metrics initialization
│   │   └── tracing.go                           # OpenTelemetry tracing setup
│   ├── db/                                      # Database initialization
│   ├── config/                                  # Configuration management
│   │   └── config.go                            # Config loading and management
│   ├── messagaging/                             # Event publishing service
│   ├── logging/                                 # Logging initialization (e.g., zap or logrus)
│   ├── jobs/                                    # Internal continuously running jobs (e.g., cron jobs)
└── └── runtime/                                 # Graceful service shudown logic
````

### 4.1 Why This Structure?

* **Maintainability:** 
  * The internal directory enforces a clear dependency flow. 
  * The domain layer is the most stable and has no dependencies on external frameworks. 
  * This makes the business logic easy to locate and reason about.

* **Testability:** 
  * Because the core logic depends on interfaces (ports), test codes can be written fast by injecting mocks.

* **Development:** 
  * A new developer can immediately understand where to add a new feature. 
  * Domain logic goes in `domain/`, a new use case in `app/`, and a new way to store data in `adapters/`. 
  * This consistency across services accelerates onboarding and development.

## 5. Testing, Security & Observability

### 5.1 Testing Strategy and Coverage
High test coverage where it matters most. 
* The business domain logic has over `90%` coverage with unit tests, ensuring all financial rules are correct. 
* The HTTP/gRPC handlers have over `80%` coverage, validating API contracts and error handling. 
* The use of interfaces makes mocking dependencies straightforward.

### 5.2 Security
* SQL injection prevention at the database layer
* JWT and RBAC at the gateway
* hashed passwords in the Auth service
* security is layered throughout the system. 
* The single API Gateway acts as a security choke point.

### 5.3 Observability
* The system is built to be transparent. Metrics (Prometheus), logs (Zerolog), and traces (OpenTelemetry) are exported, 
allowing us to monitor health, debug issues, and understand performance bottlenecks in real-time.

## 6. Installation & Deployment

The system is designed for modern deployment practices and is fully containerized.

You can run the entire system in several ways:

### 6.1 Install locally
Install on local machine local 

Prerequisite:
  * **Go** 1.24+
  * [Make](https://makefiletutorial.com/)

You can add a `.env` and update the env variables according to your need. A sample `.env.sample` file is
also provided in each service.

**Run Make Commands:** 

Go to the root folder of this project and run make commands. Make sure to run each service in a different terminal
```
# Runs Auth Service on default port (gRPC = 50051; http = 8080)
# Add .env file and add 
#  * AUTH_ENV=dev
#  * AUTH_HTTP__ADDR=:8081
#  * AUTH_GRPC__ADDR=:50051

make run-auth 

# Runs Account Service on default port (gRPC = 50051; http = 8080)
# Add .env file and add 
#  * ACCOUNT_ENV=dev
#  * ACCOUNT_HTTP__ADDR=:8082
#  * ACCOUNT_GRPC__ADDR=:50052

make run-account

# Runs Transaction Service on default port (gRPC = 50051; http = 8080)
# Add .env file and add 
#  * TRANSACTION_ENV=dev
#  * TRANSACTION_HTTP__ADDR=:8083
#  * TRANSACTION_GRPC__ADDR=:50053
#  * TRANSACTION_GRPC__ACCOUNT_SVC_ADDR=:50052

nake run-tx

# Runs Account Service on default port (gRPC = 50051; http = 8080)
# Add .env file and add 
#  * AUTH_ENV=dev
#  * GATEWAY_GRPC__AUTH_SVC_ADDR=:50051
#  * GATEWAY_GRPC__ACCOUNT_SVC_ADDR=:50052
#  * GATEWAY_GRPC__TRANSACTION_SVC_ADDR=:50053

nake run-gateway
```
### 6.2 Run on Docker
Running the system inside docker using [Docker Compose](https://docs.docker.com/compose/)

Prerequisite: 
  * [Docker](https://www.docker.com/) 
  * [Docker Compose](https://docs.docker.com/compose/)
  * [Make](https://makefiletutorial.com/)

Now run the make command
```
make compose-up
```
### 6.3 Kubernetes 

All the manifests are inside `deployment/kubernetes/manifests` folder. Update the manifests according to your cluster
configuration before applying them.

```
kubectl apply -f deployment/kubernetes/manifests/auth-service/

kubectl apply -f deployment/kubernetes/manifests/account-service/

kubectl apply -f deployment/kubernetes/manifests/transaction-service/

kubectl apply -f deployment/kubernetes/manifests/gateway-service/
```

## 7. Live Demo
You can find the live url. Go to http://bankops-core.135.235.192.122.nip.io

**To use the system, some employees with different privileges are already created. 
Use the credentials(username and password) in the `login` api for authentication and get JWT access token:** 

Employee with `admin` privileges: 
* Username: `admin`
* Password: `admin`

Employee with `editor` privileges:
* Username: `editor_user`
* Password: `editor_pass`

Employee with `viewer` privileges:
* Username: `viewer_user`
* Password: `viewer_pass`

You will get the JWT `access_token` after login. 
**Default JWT token expiry time is 15 minutes. 
This can be adjusted by updating the `AUTH_AUTH__JWT_TOKEN_DURATION` environment variable in the Auth Service**

## 8. API Documentation
Interactive API documentation, generated using Swagger/OpenAPI, is available once the Gateway service is running. 
This provides a live, explorable reference for all available endpoints, their parameters, and responses.

Access it: Just click on the **View API Documentation** button or just  go `http://<gateway-endpoint>/swagger`

## 9. Future Enhancements
The current architecture is built with future evolution in mind. 
The event-driven design and hexagonal structure make these enhancements natural next steps:

* **Distributed Caching (Redis):** New api can be added update role of employee. To improve performance and instantly reflect role changes from the Auth service, 
I plan to introduce a Redis cache. The Auth service would publish RoleUpdated events, which the Gateway would consume to invalidate its local cache.

* **Enhanced Resilience Patterns:** I've laid the groundwork for integrating circuit breakers (to prevent cascade failures) 
and a more sophisticated rate-limiting system at the Gateway.

* **SSO & Passkey Integration:** The Auth service's port-adapter design makes it trivial to add new authentication providers 
like Single Sign-On (SSO) or modern passkeys by simply writing a new adapter.