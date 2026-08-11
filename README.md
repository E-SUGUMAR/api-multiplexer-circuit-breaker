# API Multiplexer with Circuit Breaker

A fault-tolerant API multiplexer built using **Go, React, Docker, and Toxiproxy**.

The system routes requests through a Primary API and automatically falls back to a Secondary API when the Primary API becomes slow or unavailable. A Circuit Breaker prevents repeated requests to the failing Primary API and allows recovery testing after a configured timeout.

## Architecture

```text
                    ┌─────────────────┐
                    │  React Frontend │
                    │   Port: 5173    │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   Go Backend    │
                    │   Port: 8080    │
                    │                 │
                    │ Circuit Breaker │
                    └───────┬─────────┘
                            │
                 ┌──────────┴──────────┐
                 │                     │
                 ▼                     ▼
        ┌─────────────────┐   ┌─────────────────┐
        │   Toxiproxy     │   │ Secondary API   │
        │    :8666        │   │    :8082        │
        └────────┬────────┘   └─────────────────┘
                 │
                 ▼
        ┌─────────────────┐
        │   Primary API   │
        │    :8081        │
        └─────────────────┘

Request Flow

Under normal conditions:

    Frontend
       ↓
    Go Backend
       ↓
    Toxiproxy
       ↓
    Primary API

When the Primary API fails or exceeds the timeout:

    Frontend
       ↓
    Go Backend
       ↓
    Primary fails
       ↓
    Circuit Breaker records failure
       ↓
    Secondary API

When the Circuit Breaker is OPEN:

    Frontend
       ↓
    Go Backend
       ↓
    Circuit OPEN
       ↓
    Secondary API

Important: Toxiproxy is used only for the Primary API. The Secondary API is accessed directly and does not pass through Toxiproxy.

Technologies Used
   1. Go - API multiplexer and circuit breaker
   2. React - Monitoring dashboard
   3. Docker - Containerization
   4. Docker Compose - Multi-container orchestration
   5. Toxiproxy - Network fault and latency injection
   6. Node.js / npm - React frontend
   7. PowerShell / curl - API and Toxiproxy testing

Project Structure

ForntEnd_Project/
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
│
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── package-lock.json
│   ├── src/
│   └── public/
│
├── primary-api/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
│
├── secondary-api/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
│
├── docker-compose.yml
├── .gitignore
└── README.md

Configuration

The current backend configuration uses:

| Configuration           |        Value |
| ----------------------- | -----------: |
| Backend                 |       `8080` |
| Primary API             |       `8081` |
| Secondary API           |       `8082` |
| Toxiproxy API           |       `8474` |
| Toxiproxy Primary Proxy |       `8666` |
| Primary timeout         |     `200 ms` |
| Failure threshold       | `3` failures |
| Recovery timeout        |  `5 seconds` |
| Test latency            |     `500 ms` |

Circuit Breaker

The backend uses three circuit breaker states.

1. CLOSED

Normal operating state.

Requests are sent to the Primary API through Toxiproxy.

    Backend
       ↓
    Toxiproxy
       ↓
    Primary API

A successful Primary request resets the failure counter.

2. OPEN

The Circuit Breaker enters the OPEN state after 3 Primary failures.

Once OPEN, requests do not attempt the Primary API.

Instead:

    Backend
       ↓
    Secondary API

This prevents repeatedly sending requests to a failing Primary API.

3. HALF-OPEN

After the configured 5-second recovery period, the Circuit Breaker allows a request to test the Primary API again.

If the Primary succeeds:

    HALF-OPEN
       ↓
    Primary succeeds
       ↓
    CLOSED

If the Primary fails:

    HALF-OPEN
        ↓
    Primary fails
        ↓
    OPEN

Toxiproxy

Toxiproxy is used to simulate network problems for the Primary API.

The Primary route is:

    Go Backend
        ↓
    Toxiproxy :8666
        ↓
    Primary API :8081

The Secondary API does not use Toxiproxy:

    Go Backend
        ↓
    Secondary API :8082

500 ms Latency

A 500 ms latency toxic can be applied to the Primary proxy.

Because the backend Primary request timeout is configured to:

200 ms

the Primary request can exceed the timeout and be treated as a failure.

This allows the circuit breaker and fallback mechanism to be tested.

Running the Project

Make sure Docker Desktop is running.

Open a terminal in the project root:

cd "E:\Data_Science\Data_Support_Intern\Intern_ML_Task\Course_Videos\ForntEnd_Project"

Start the complete application:

docker compose up --build

The following services are started:

primary
secondary
toxiproxy
toxiproxy-init
backend
frontend

Check Containers

Open another terminal in the project directory and run:

docker compose ps

You should see the application containers running.

View Backend Logs

Run:

docker compose logs -f backend

The backend displays information about:

   1. Primary requests
   2. Primary failures
   3. Circuit Breaker state
   4. Secondary fallback
   5. Routing decisions

Example:

Routing request → Primary through Toxiproxy
Primary Failed
Primary failure count: 1/3
Primary failed → Routing directly to Secondary
Secondary request successful

After three failures:

Primary failure count: 3/3
Circuit Breaker → OPEN
Circuit OPEN → Routing directly to Secondary

Test the Backend

The backend is available at:

http://localhost:8080

Using PowerShell:

curl.exe http://localhost:8080

When the Primary API is healthy, the response should come from the Primary API.

When the Primary API is delayed beyond the configured timeout, the backend falls back to the Secondary API.

Check Circuit Breaker Status

The backend exposes:

http://localhost:8080/status

Test it with:

curl.exe http://localhost:8080/status

The response contains:

{
  "circuit": "CLOSED",
  "activeRoute": "PRIMARY",
  "requests": 10,
  "rps": 1
}

Possible circuit states are:

CLOSED
OPEN
HALF-OPEN

Possible active routes are:

PRIMARY
SECONDARY

Test Toxiproxy

Check the configured Primary proxy:

curl.exe http://localhost:8474/proxies

The Primary proxy should point to:

listen: 8666
upstream: primary:8081

When latency is configured, the proxy should contain a latency toxic similar to:

latency: 500 ms

Fault Injection Test

To test the circuit breaker:

Step 1 - Enable 500 ms latency

Configure a 500 ms latency toxic on the Primary Toxiproxy route.

Step 2 - Send requests

Run:

curl.exe http://localhost:8080

multiple times.

Step 3 - Observe the backend logs

The Primary request should exceed the 200 ms timeout.

The backend should record Primary failures:

Primary failure count: 1/3
Primary failure count: 2/3
Primary failure count: 3/3

After the third failure:

Circuit Breaker → OPEN

The backend then routes requests directly to:

Secondary API :8082
Step 4 - Wait for recovery

After approximately 5 seconds, the Circuit Breaker enters:

HALF-OPEN

A Primary request is then used to test whether the Primary API has recovered.

Frontend Dashboard

The React frontend runs on:

http://localhost:5173

The dashboard communicates with the Go backend and displays information such as:

Circuit Breaker state
Active route
Request count
Requests per second

The frontend obtains status information from:

http://localhost:8080/status

Docker Services

The Docker Compose setup contains:

| Service          | Purpose                         |           Port |
| ---------------- | ------------------------------- | -------------: |
| `frontend`       | React dashboard                 |         `5173` |
| `backend`        | Go API multiplexer              |         `8080` |
| `primary`        | Primary API                     |         `8081` |
| `secondary`      | Secondary API                   |         `8082` |
| `toxiproxy`      | Primary fault injection         | `8474`, `8666` |
| `toxiproxy-init` | Initial Toxiproxy configuration |              - |

Stopping the Application

To stop the running Docker Compose application:

docker compose down

To stop and remove the containers:

docker compose down

To rebuild the project after making code changes:

docker compose up --build

GitHub

Repository:

api-multiplexer-circuit-breaker

This project demonstrates API routing, fault tolerance, circuit breaker behavior, network latency simulation, automatic failover, and containerized application deployment.

Future Improvements

Possible future improvements include:

    1. More detailed monitoring metrics
    2. Retry policies
    3. Configurable circuit breaker parameters
    4. Health-check endpoints
    5. Prometheus metrics
    6. Grafana monitoring
    7. Additional Toxiproxy failure scenarios
    8. Improved frontend visualization