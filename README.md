# 🚀 Fault-Tolerant API Multiplexer

A production-style **fault-tolerant API routing system** built using **Go, React, WebSocket, Docker Compose, Toxiproxy, and Circuit Breaker patterns**.

The system routes requests through a Primary API under normal conditions and automatically falls back to a Secondary API when the Primary API becomes slow or unreliable.

It also provides a real-time React dashboard using WebSocket communication to monitor the circuit state, active route, request count, and system performance.

---

# 📌 Project Overview

The project demonstrates how to build a resilient API architecture capable of handling:

* Primary API failures
* Network latency
* Partial failures
* Automatic failover
* Circuit breaker state transitions
* Service recovery
* Real-time monitoring
* Resource-constrained backend execution

The architecture uses **Toxiproxy** to introduce controlled network failures and latency without modifying the Primary API itself.

---

# 🏗️ Architecture

```text
                         ┌──────────────────────┐
                         │    React Dashboard   │
                         │      Port 5173       │
                         └──────────┬───────────┘
                                    │
                              WebSocket
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │     Go Multiplexer    │
                         │      Port 8080        │
                         │                      │
                         │   Circuit Breaker    │
                         └──────────┬───────────┘
                                    │
                         ┌──────────┴──────────┐
                         │                     │
                         ▼                     ▼
                ┌──────────────────┐   ┌──────────────────┐
                │    Toxiproxy     │   │  Secondary API   │
                │    Port 8666     │   │    Port 8082     │
                └────────┬─────────┘   └──────────────────┘
                         │
                         ▼
                ┌──────────────────┐
                │    Primary API   │
                │    Port 8081     │
                └──────────────────┘
```

---

# 🧩 Components

| Component       | Technology  |   Port | Purpose                     |
| --------------- | ----------- | -----: | --------------------------- |
| Primary API     | API Service | `8081` | Main request handler        |
| Secondary API   | API Service | `8082` | Backup/fallback service     |
| API Multiplexer | Go          | `8080` | Routing and circuit breaker |
| Toxiproxy       | Toxiproxy   | `8666` | Primary API network proxy   |
| Toxiproxy API   | Toxiproxy   | `8474` | Chaos configuration         |
| Dashboard       | React       | `5173` | Real-time monitoring        |
| Communication   | WebSocket   |      — | Real-time dashboard updates |

---

# ✅ Completed Features

The project successfully implements all required functionality:

1. ✅ Primary API
2. ✅ Secondary API
3. ✅ Go API Multiplexer
4. ✅ Circuit Breaker
5. ✅ `CLOSED → OPEN → HALF-OPEN → CLOSED`
6. ✅ Primary API through Toxiproxy
7. ✅ Secondary API direct
8. ✅ 500 ms latency chaos test
9. ✅ 20% failure scenario
10. ✅ Automatic fallback
11. ✅ Automatic recovery
12. ✅ React dashboard
13. ✅ WebSocket communication
14. ✅ Docker Compose
15. ✅ 128 MB backend memory limit
16. ✅ Final cleanup
17. ✅ End-to-end fault-tolerance testing

---

# 🔀 Request Routing

## Normal Operation

Under normal conditions, requests follow:

```text
Client
  │
  ▼
Go API Multiplexer
  │
  ▼
Circuit Breaker
  │
  ▼
Toxiproxy
  │
  ▼
Primary API
```

The Primary API is the preferred route.

```text
Active Route: PRIMARY
Circuit: CLOSED
```

---

# 🔥 Circuit Breaker

The Go backend implements a three-state Circuit Breaker:

```text
             ┌───────────────┐
             │    CLOSED     │
             └───────┬───────┘
                     │
              Failure threshold
                     │
                     ▼
             ┌───────────────┐
             │     OPEN      │
             └───────┬───────┘
                     │
               Recovery time
                     │
                     ▼
             ┌───────────────┐
             │   HALF-OPEN   │
             └───────┬───────┘
                     │
              Test Primary API
                │          │
             Success      Failure
                │          │
                ▼          ▼
             CLOSED       OPEN
```

---

# 🟢 CLOSED State

The system starts in:

```text
CLOSED
```

Requests are routed to:

```text
Primary API
```

Traffic flow:

```text
Go Backend
     │
     ▼
Toxiproxy
     │
     ▼
Primary API
```

---

# 🔴 OPEN State

When the Primary API continuously fails or exceeds the configured timeout, the failure counter increases.

After reaching the configured failure threshold:

```text
CLOSED
   ↓
OPEN
```

The Primary API is temporarily bypassed.

Requests are immediately sent to:

```text
Secondary API
```

Traffic flow:

```text
Go Backend
     │
     ▼
Secondary API
```

This prevents repeated requests from continuously hitting an unhealthy Primary API.

---

# 🟡 HALF-OPEN State

After the configured recovery period, the Circuit Breaker enters:

```text
HALF-OPEN
```

The system allows a test request to determine whether the Primary API has recovered.

### Successful Test

```text
HALF-OPEN
    ↓
Primary API successful
    ↓
CLOSED
```

Traffic returns to the Primary API.

### Failed Test

```text
HALF-OPEN
    ↓
Primary API failed
    ↓
OPEN
```

Traffic continues through the Secondary API.

---

# 🌐 Toxiproxy

The Primary API is accessed through Toxiproxy.

```text
Go Backend
     │
     ▼
Toxiproxy :8666
     │
     ▼
Primary API :8081
```

The Secondary API does **not** use Toxiproxy.

```text
Go Backend
     │
     ▼
Secondary API :8082
```

This allows the Primary API to be intentionally degraded while keeping the Secondary API healthy.

---

# 🌪️ Chaos Testing

Toxiproxy is used to simulate real-world network problems.

Two important chaos scenarios were implemented.

---

## 1. 500 ms Latency Test

A `500 ms` latency condition is introduced between the Go backend and Primary API.

```text
Request
   │
   ▼
Go Backend
   │
   ▼
Toxiproxy
   │
   │ 500 ms delay
   ▼
Primary API
```

The latency causes the Primary API request to exceed the backend timeout.

The Circuit Breaker detects the failure and eventually switches traffic to the Secondary API.

---

# 📉 20% Failure Scenario

A separate chaos scenario introduces approximately:

```text
20% failure rate
```

This simulates intermittent Primary API failures rather than a complete outage.

The purpose is to verify that the multiplexer can handle partial/unreliable failures while maintaining service availability through the fallback mechanism.

---

# 🔄 Automatic Fallback

When the Primary API becomes unhealthy:

```text
Primary API
     │
     X
     │
     ▼
Circuit Breaker
     │
     ▼
Secondary API
```

No manual route change is required.

The Go multiplexer automatically changes the active route.

Example:

```text
Active Route: PRIMARY
```

changes to:

```text
Active Route: SECONDARY
```

---

# ♻️ Automatic Recovery

Once the Primary API becomes healthy again:

```text
OPEN
  ↓
HALF-OPEN
  ↓
Primary health test
  ↓
Successful
  ↓
CLOSED
  ↓
PRIMARY
```

The system automatically returns traffic to the Primary API.

No restart is required.

---

# 🖥️ React Dashboard

The React dashboard provides real-time visibility into the system.

The dashboard displays information such as:

* Circuit Breaker state
* Active API route
* Total request count
* Requests per second
* Current system status

Example:

```text
┌─────────────────────────────────────────┐
│          API MULTIPLEXER                │
├─────────────────────────────────────────┤
│                                         │
│ Circuit:       CLOSED                   │
│ Active Route:  PRIMARY                  │
│ Requests:      1250                     │
│ RPS:           10                       │
│                                         │
└─────────────────────────────────────────┘
```

---

# ⚡ WebSocket Communication

The dashboard uses **WebSocket communication** for real-time status updates.

Instead of repeatedly polling the backend for status information, the backend can push updated information to connected clients.

```text
Go Backend
     │
     │ WebSocket
     ▼
React Dashboard
```

This allows the dashboard to immediately reflect:

```text
CLOSED
   ↓
OPEN
   ↓
HALF-OPEN
   ↓
CLOSED
```

and route changes:

```text
PRIMARY
   ↓
SECONDARY
   ↓
PRIMARY
```

---

# 🐳 Docker Compose

All application components are containerized and managed using Docker Compose.

The project contains separate services for:

```text
Primary API
Secondary API
Go Backend
React Frontend
Toxiproxy
```

The entire environment can be started together using:

```powershell
docker compose up --build
```

---

# 💾 Backend Memory Limit

The Go backend container is configured with a memory limit of:

```text
128 MB
```

This demonstrates that the backend can operate under a controlled resource constraint.

The resource configuration is managed through Docker Compose.

---

# 📁 Project Structure

```text
FrontEnd_Project/
│
├── backend/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── App.jsx
│   │   ├── App.css
│   │   └── ...
│   ├── package.json
│   └── Dockerfile
│
├── primary-api/
│   ├── ...
│   └── Dockerfile
│
├── secondary-api/
│   ├── ...
│   └── Dockerfile
│
├── docker-compose.yml
└── README.md
```

---

# ▶️ Running the Project

## 1. Start the Complete Application

From the project root:

```powershell
docker compose up --build
```

---

## 2. Run in Background

```powershell
docker compose up -d
```

---

## 3. Check Containers

```powershell
docker compose ps
```

---

## 4. Open the React Dashboard

```text
http://localhost:5173
```

---

## 5. Check Backend

```text
http://localhost:8080
```

---

## 6. Check Backend Status

```text
http://localhost:8080/status
```

---

# 🧪 Testing

## Test Normal Operation

Send a request:

```powershell
curl.exe http://localhost:8080
```

Expected behavior:

```text
Backend
   ↓
Toxiproxy
   ↓
Primary API
```

Circuit:

```text
CLOSED
```

Route:

```text
PRIMARY
```

---

# 🧪 Test 500 ms Latency

Add latency through Toxiproxy:

```powershell
curl.exe -X POST "http://localhost:8474/proxies/primary/toxics" -H "Content-Type: application/json" -d '{"name":"latency","type":"latency","attributes":{"latency":500}}'
```

Verify:

```powershell
curl.exe http://localhost:8474/proxies
```

The Primary proxy should contain the latency toxic.

---

# 🧪 Test Failover

Send multiple requests:

```powershell
curl.exe http://localhost:8080
```

The Primary API should begin failing because of the injected latency.

The Circuit Breaker eventually transitions:

```text
CLOSED
   ↓
OPEN
```

The backend automatically routes requests to:

```text
SECONDARY
```

---

# 🧪 Test Recovery

Remove the latency toxic:

```powershell
curl.exe -X DELETE "http://localhost:8474/proxies/primary/toxics/latency"
```

Verify:

```powershell
curl.exe http://localhost:8474/proxies
```

After the recovery period:

```text
OPEN
   ↓
HALF-OPEN
   ↓
Primary test successful
   ↓
CLOSED
```

Traffic returns to:

```text
PRIMARY
```

---

# 🧹 Cleanup

Stop all containers:

```powershell
docker compose down
```

Remove containers and associated resources when required:

```powershell
docker compose down --remove-orphans
```

The project has been cleaned up after testing so that the final environment contains only the required configuration.

---

# 📊 Final System Behavior

The complete fault-tolerance cycle is:

```text
                  ┌──────────────┐
                  │   PRIMARY    │
                  │    HEALTHY   │
                  └──────┬───────┘
                         │
                         ▼
                     CLOSED
                         │
                 Primary failure
                         │
                         ▼
                       OPEN
                         │
                 Automatic fallback
                         │
                         ▼
                    SECONDARY
                         │
                 Recovery period
                         │
                         ▼
                    HALF-OPEN
                         │
                  Test Primary
                    /       \
                   /         \
              Success       Failure
                 │             │
                 ▼             ▼
              CLOSED          OPEN
                 │
                 ▼
              PRIMARY
```

---

# 🎯 Project Goals Achieved

| Requirement                        | Status |
| ---------------------------------- | :----: |
| Primary API                        |    ✅   |
| Secondary API                      |    ✅   |
| Go API Multiplexer                 |    ✅   |
| Circuit Breaker                    |    ✅   |
| CLOSED → OPEN → HALF-OPEN → CLOSED |    ✅   |
| Primary through Toxiproxy          |    ✅   |
| Secondary direct                   |    ✅   |
| 500 ms latency chaos test          |    ✅   |
| 20% failure scenario               |    ✅   |
| Automatic fallback                 |    ✅   |
| Automatic recovery                 |    ✅   |
| React dashboard                    |    ✅   |
| WebSocket                          |    ✅   |
| Docker Compose                     |    ✅   |
| 128 MB backend memory limit        |    ✅   |
| Final cleanup                      |    ✅   |

---

# 🏁 Conclusion

This project demonstrates a complete **fault-tolerant API architecture** with automatic failure detection, intelligent request routing, network chaos testing, service fallback, and automatic recovery.

The combination of:

```text
Go
+
Circuit Breaker
+
Toxiproxy
+
Primary API
+
Secondary API
+
React
+
WebSocket
+
Docker Compose
```

provides a practical demonstration of building resilient backend systems capable of maintaining service availability even when the Primary API experiences latency or failures.

---

## 🔑 Key Takeaway

The system follows the principle:

```text
Healthy Primary
      ↓
Use Primary

Primary Unhealthy
      ↓
Open Circuit
      ↓
Use Secondary

Primary Recovers
      ↓
Half-Open Test
      ↓
Return to Primary
```

This completes the implementation and testing of the fault-tolerant API multiplexer.
