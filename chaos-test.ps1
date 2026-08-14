# ============================================================
# API Multiplexer - Chaos Test
# Toxiproxy chaos testing for Primary API
#
# Tests:
#   1. 500ms latency
#   2. 20% connection-failure simulation
# ============================================================

Write-Host "========================================"
Write-Host "API Multiplexer Chaos Test"
Write-Host "========================================"

$toxiproxy = "http://localhost:8474"
$proxy = "primary"

# ============================================================
# CLEAN EXISTING TOXICS
# ============================================================

Write-Host "`nCleaning existing toxics..."

curl.exe -X DELETE `
    "$toxiproxy/proxies/$proxy/toxics/latency" `
    -H "User-Agent: curl"

curl.exe -X DELETE `
    "$toxiproxy/proxies/$proxy/toxics/connection_failure" `
    -H "User-Agent: curl"

# ============================================================
# TEST 1 - 500ms LATENCY
# ============================================================

Write-Host "`n========================================"
Write-Host "TEST 1: Adding 500ms latency"
Write-Host "========================================"

$latencyBody = '{"name":"latency","type":"latency","attributes":{"latency":500}}'

curl.exe -X POST `
    "$toxiproxy/proxies/$proxy/toxics" `
    -H "Content-Type: application/json" `
    -H "User-Agent: curl" `
    --data-binary $latencyBody

Write-Host "`nCurrent Toxiproxy configuration:"

curl.exe "$toxiproxy/proxies"

# ============================================================
# TEST LATENCY
# ============================================================

Write-Host "`n========================================"
Write-Host "Testing latency through Backend"
Write-Host "========================================"

for ($i = 1; $i -le 5; $i++) {

    Write-Host "`nRequest $i"

    curl.exe http://localhost:8080/
}

Write-Host "`nChecking circuit state..."

curl.exe http://localhost:8080/status

# ============================================================
# REMOVE LATENCY
# ============================================================

Write-Host "`n========================================"
Write-Host "Removing 500ms latency"
Write-Host "========================================"

curl.exe -X DELETE `
    "$toxiproxy/proxies/$proxy/toxics/latency" `
    -H "User-Agent: curl"

Start-Sleep -Seconds 6

# ============================================================
# TEST RECOVERY
# ============================================================

Write-Host "`n========================================"
Write-Host "Testing Primary after recovery"
Write-Host "========================================"

curl.exe http://localhost:8080/

Write-Host "`nCircuit status after recovery..."

curl.exe http://localhost:8080/status

# ============================================================
# TEST 2 - 20% CONNECTION FAILURE
# ============================================================

Write-Host "`n========================================"
Write-Host "TEST 2: 20% connection failure"
Write-Host "========================================"

$failureBody = '{"name":"connection_failure","type":"timeout","attributes":{"timeout":1000},"toxicity":0.2}'

curl.exe -X POST `
    "$toxiproxy/proxies/$proxy/toxics" `
    -H "Content-Type: application/json" `
    -H "User-Agent: curl" `
    --data-binary $failureBody

Write-Host "`nCurrent Toxiproxy configuration:"

curl.exe "$toxiproxy/proxies"

# ============================================================
# TEST FAILURE SIMULATION
# ============================================================

Write-Host "`n========================================"
Write-Host "Testing Primary multiple times"
Write-Host "========================================"

for ($i = 1; $i -le 30; $i++) {

    Write-Host "`nRequest $i"

    curl.exe http://localhost:8080/
}

# ============================================================
# CHECK STATUS
# ============================================================

Write-Host "`n========================================"
Write-Host "Final Circuit Status"
Write-Host "========================================"

curl.exe http://localhost:8080/status

# ============================================================
# BACKEND LOGS
# ============================================================

Write-Host "`n========================================"
Write-Host "Chaos test completed"
Write-Host "========================================"

Write-Host "`nFinal Toxiproxy configuration:"

curl.exe "$toxiproxy/proxies"