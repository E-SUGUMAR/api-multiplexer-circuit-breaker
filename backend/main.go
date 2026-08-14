package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// ============================================================
	// PRIMARY API
	// ============================================================
	//
	// IMPORTANT:
	// Primary traffic goes THROUGH TOXIPROXY.
	//
	// Go Backend
	//     ↓
	// Toxiproxy :8666
	//     ↓
	// Primary API :8081
	//
	PRIMARY_URL = "http://toxiproxy:8666" //"http://localhost:8666"

	// ============================================================
	// SECONDARY API
	// ============================================================
	//
	// IMPORTANT:
	// Secondary does NOT use Toxiproxy.
	//
	// Go Backend
	//     ↓
	// Secondary API :8082
	//
	SECONDARY_URL = "http://secondary:8082" //"http://localhost:8082"

	// ============================================================
	// CIRCUIT BREAKER / REQUEST CONFIGURATION
	// ============================================================

	// Assessment requirement:
	// Primary request must timeout after exactly 200 ms.
	REQUEST_TIMEOUT = 200 * time.Millisecond

	// Number of Primary failures required to OPEN circuit.
	FAILURE_THRESHOLD = 3

	// Time the circuit remains OPEN before testing Primary again.
	RECOVERY_TIMEOUT = 5 * time.Second
)

// ================================================================
// CIRCUIT BREAKER
// ================================================================

type CircuitBreaker struct {
	mu sync.Mutex

	state string

	failureCount int

	lastFailure time.Time
}

var cb = &CircuitBreaker{
	state: "CLOSED",
}

// ================================================================
// GLOBAL METRICS
// ================================================================

// atomic.Int64 is used because these values are accessed
// concurrently by multiple HTTP requests and the RPS goroutine.

var requestCount atomic.Int64

var requestsPerSecond atomic.Int64

// activeRoute is accessed concurrently, therefore atomic.Value
// is used instead of a normal string.
var activeRoute atomic.Value

func init() {
	activeRoute.Store("PRIMARY")
}

// ================================================================
// WEBSOCKET TELEMETRY
// ================================================================

type Metrics struct {
	Circuit     string `json:"circuit"`
	ActiveRoute string `json:"activeRoute"`
	Requests    int64  `json:"requests"`
	RPS         int64  `json:"rps"`
}

var wsClients = make(map[*websocket.Conn]bool)

var wsMutex sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow the React frontend running on localhost:5173.
		return true
	},
}

// getMetrics returns the current dashboard metrics.
func getMetrics() Metrics {

	return Metrics{
		Circuit:     cb.getState(),
		ActiveRoute: activeRoute.Load().(string),
		Requests:    requestCount.Load(),
		RPS:         requestsPerSecond.Load(),
	}
}

// addWebSocketClient adds a browser WebSocket connection.
func addWebSocketClient(conn *websocket.Conn) {

	wsMutex.Lock()
	defer wsMutex.Unlock()

	wsClients[conn] = true
}

// removeWebSocketClient removes a disconnected browser.
func removeWebSocketClient(conn *websocket.Conn) {

	wsMutex.Lock()
	defer wsMutex.Unlock()

	delete(wsClients, conn)
}

// broadcastMetrics sends the latest metrics
// to every connected frontend.
func broadcastMetrics() {

	metrics := getMetrics()

	data, err := json.Marshal(metrics)

	if err != nil {
		fmt.Println("WebSocket JSON error:", err)
		return
	}

	wsMutex.Lock()
	defer wsMutex.Unlock()

	for conn := range wsClients {

		conn.SetWriteDeadline(
			time.Now().Add(1 * time.Second),
		)

		err := conn.WriteMessage(
			websocket.TextMessage,
			data,
		)

		if err != nil {

			fmt.Println(
				"WebSocket client disconnected:",
				err,
			)

			conn.Close()

			delete(wsClients, conn)
		}
	}
}

// websocketHandler upgrades the HTTP connection
// to a WebSocket connection.
func websocketHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	conn, err := upgrader.Upgrade(
		w,
		r,
		nil,
	)

	if err != nil {

		fmt.Println(
			"WebSocket upgrade failed:",
			err,
		)

		return
	}

	fmt.Println(
		"WebSocket client connected",
	)

	addWebSocketClient(conn)

	defer func() {

		removeWebSocketClient(conn)

		conn.Close()

		fmt.Println(
			"WebSocket client disconnected",
		)

	}()

	// Keep reading from the browser so that
	// WebSocket close/ping messages are handled.
	for {

		_, _, err := conn.ReadMessage()

		if err != nil {
			return
		}
	}
}

// startWebSocketBroadcaster continuously sends
// telemetry to connected clients.
func startWebSocketBroadcaster() {

	ticker := time.NewTicker(
		100 * time.Millisecond,
	)

	defer ticker.Stop()

	for range ticker.C {

		broadcastMetrics()
	}
}

// ================================================================
// CIRCUIT BREAKER - CHECK FALLBACK
// ================================================================

func (cb *CircuitBreaker) shouldUseFallback() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "OPEN" {
		if time.Since(cb.lastFailure) >= RECOVERY_TIMEOUT {
			cb.state = "HALF-OPEN"

			fmt.Println("Circuit OPEN timeout expired → HALF-OPEN")

			// Allow a request to test Primary.
			return false
		}

		return true
	}

	if cb.state == "HALF-OPEN" {
		// For this project, don't allow another request
		// to continuously test Primary while one test is in progress.
		return true
	}

	// CLOSED
	return false
}

// ================================================================
// CIRCUIT BREAKER - SUCCESS
// ================================================================

func (cb *CircuitBreaker) success() {

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// A successful Primary request closes the circuit
	// and resets the failure counter.

	cb.failureCount = 0

	cb.state = "CLOSED"
	fmt.Println("Primary request successful → Circuit Breaker → CLOSED")
}

// ================================================================
// CIRCUIT BREAKER - FAILURE
// ================================================================

func (cb *CircuitBreaker) failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// If Primary fails while testing recovery,
	// immediately return the circuit to OPEN.
	if cb.state == "HALF-OPEN" {
		cb.state = "OPEN"
		cb.lastFailure = time.Now()

		fmt.Println("HALF-OPEN Primary test failed → Circuit OPEN")
		return
	}

	cb.failureCount++

	fmt.Printf(
		"Primary failure count: %d/%d\n",
		cb.failureCount,
		FAILURE_THRESHOLD,
	)

	if cb.failureCount >= FAILURE_THRESHOLD {
		cb.state = "OPEN"
		cb.lastFailure = time.Now()

		fmt.Println("Circuit Breaker → OPEN")
	}
}

// ================================================================
// GET CIRCUIT STATE
// ================================================================

func (cb *CircuitBreaker) getState() string {

	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.state
}

// ================================================================
// REVERSE PROXY
// ================================================================

func reverseProxy(
	target string,
	w http.ResponseWriter,
	r *http.Request,
) error {

	// ------------------------------------------------------------
	// 200 ms context timeout
	// ------------------------------------------------------------

	ctx, cancel := context.WithTimeout(
		r.Context(),
		REQUEST_TIMEOUT,
	)

	defer cancel()

	// ------------------------------------------------------------
	// Build target URL
	// ------------------------------------------------------------

	targetURL := target + r.URL.Path

	// Preserve query parameters.

	if r.URL.RawQuery != "" {

		targetURL += "?" + r.URL.RawQuery
	}

	// ------------------------------------------------------------
	// Create downstream request
	// ------------------------------------------------------------

	req, err := http.NewRequestWithContext(
		ctx,
		r.Method,
		targetURL,
		r.Body,
	)

	if err != nil {
		return err
	}

	// Copy incoming headers.

	req.Header = r.Header.Clone()

	// ------------------------------------------------------------
	// HTTP client
	// ------------------------------------------------------------

	resp, err := http.DefaultClient.Do(req)

	if err != nil {

		// This includes:
		//
		// context deadline exceeded
		// connection refused
		// connection reset
		// etc.

		return err
	}

	defer resp.Body.Close()

	// ------------------------------------------------------------
	// Copy response headers
	// ------------------------------------------------------------

	for key, values := range resp.Header {

		for _, value := range values {

			w.Header().Add(
				key,
				value,
			)
		}
	}

	// ------------------------------------------------------------
	// Send response status
	// ------------------------------------------------------------

	w.WriteHeader(resp.StatusCode)

	// ------------------------------------------------------------
	// Copy response body
	// ------------------------------------------------------------

	_, err = io.Copy(
		w,
		resp.Body,
	)

	return err
}

// ================================================================
// MAIN ROUTER
// ================================================================

func router(
	w http.ResponseWriter,
	r *http.Request,
) {

	requestCount.Add(1)

	// ============================================================
	// CHECK CIRCUIT BREAKER
	// ============================================================

	if cb.shouldUseFallback() {

		// Circuit is OPEN.
		// Directly use Secondary.

		activeRoute.Store("SECONDARY")

		fmt.Println(
			"Circuit OPEN → Routing directly to Secondary",
		)

		err := reverseProxy(
			SECONDARY_URL,
			w,
			r,
		)

		if err != nil {

			fmt.Println(
				"Secondary failed:",
				err,
			)

			http.Error(
				w,
				"Secondary API unavailable",
				http.StatusServiceUnavailable,
			)
		}

		return
	}

	// ============================================================
	// TRY PRIMARY
	// ============================================================

	activeRoute.Store("PRIMARY")

	fmt.Println(
		"Routing request → Primary through Toxiproxy",
	)

	// PRIMARY_URL is:
	//
	// http://localhost:8666
	//
	// NOT:
	//
	// http://localhost:8081
	//
	// This means Primary traffic goes through Toxiproxy.

	err := reverseProxy(
		PRIMARY_URL,
		w,
		r,
	)

	// ============================================================
	// PRIMARY SUCCESS
	// ============================================================

	if err == nil {

		fmt.Println(
			"Primary request successful",
		)

		cb.success()

		return
	}

	// ============================================================
	// PRIMARY FAILURE
	// ============================================================

	fmt.Println(
		"Primary Failed:",
		err,
	)

	cb.failure()

	// ============================================================
	// FALLBACK TO SECONDARY
	// ============================================================

	activeRoute.Store("SECONDARY")

	fmt.Println(
		"Primary failed → Routing directly to Secondary",
	)

	err = reverseProxy(
		SECONDARY_URL,
		w,
		r,
	)

	// ============================================================
	// SECONDARY FAILURE
	// ============================================================

	if err != nil {

		fmt.Println(
			"Secondary Failed:",
			err,
		)

		http.Error(
			w,
			"Both Primary and Secondary APIs are unavailable",
			http.StatusServiceUnavailable,
		)

		return
	}

	fmt.Println(
		"Secondary request successful",
	)
}

// ================================================================
// STATUS ENDPOINT
// ================================================================

func statusHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	circuitState := cb.getState()

	route := activeRoute.Load().(string)

	requests := requestCount.Load()

	rps := requestsPerSecond.Load()

	fmt.Fprintf(
		w,
		`{
		"circuit":"%s",
		"activeRoute":"%s",
		"requests":%d,
		"rps":%d
	}`,
		circuitState,
		route,
		requests,
		rps,
	)
}

// ================================================================
// CORS
// ================================================================

func enableCORS(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			w.Header().Set(
				"Access-Control-Allow-Origin",
				"http://localhost:5173",
			)

			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, OPTIONS",
			)

			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type",
			)

			// Handle browser preflight.

			if r.Method == "OPTIONS" {

				w.WriteHeader(
					http.StatusOK,
				)

				return
			}

			next.ServeHTTP(
				w,
				r,
			)
		},
	)
}

// ================================================================
// REQUESTS PER SECOND
// ================================================================

func calculateRPS() {

	ticker := time.NewTicker(
		1 * time.Second,
	)

	defer ticker.Stop()

	var lastCount int64

	for range ticker.C {

		currentCount := requestCount.Load()

		rps := currentCount - lastCount

		requestsPerSecond.Store(rps)

		lastCount = currentCount
	}
}

// ================================================================
// MAIN
// ================================================================

func main() {

	// ------------------------------------------------------------
	// Routes
	// ------------------------------------------------------------

	http.HandleFunc(
		"/",
		router,
	)

	http.HandleFunc(
		"/status",
		statusHandler,
	)

	http.HandleFunc(
		"/ws",
		websocketHandler,
	)

	// ------------------------------------------------------------
	// Start RPS calculation
	// ------------------------------------------------------------

	go calculateRPS()

	// ------------------------------------------------------------
	// Start WebSocket telemetry
	// ------------------------------------------------------------

	go startWebSocketBroadcaster()

	// ------------------------------------------------------------
	// Start backend
	// ------------------------------------------------------------

	fmt.Println(
		"========================================",
	)

	fmt.Println(
		"Go API Multiplexer",
	)

	fmt.Println(
		"========================================",
	)

	fmt.Println(
		"Backend       : http://localhost:8080",
	)

	fmt.Println(
		"Primary       : Toxiproxy :8666 → :8081",
	)

	fmt.Println(
		"Secondary     : Direct → :8082",
	)

	fmt.Println(
		"Timeout       : 200 ms",
	)

	fmt.Println(
		"Failure limit : 3",
	)

	fmt.Println(
		"Recovery      : 5 seconds",
	)

	fmt.Println(
		"========================================",
	)

	handler := enableCORS(
		http.DefaultServeMux,
	)

	// ------------------------------------------------------------
	// Start HTTP server
	// ------------------------------------------------------------

	log.Fatal(
		http.ListenAndServe(
			":8080",
			handler,
		),
	)
}
