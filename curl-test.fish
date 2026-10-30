#!/usr/bin/env fish

set -l base_url "http://localhost:8080"
set -l endpoints /test /delay /flaky /payload /unknown
set -l total_requests 100

echo "=== Checking connectivity to $base_url ==="
if not curl -s --connect-timeout 1 "$base_url/test" >/dev/null
    echo "Error: $base_url is unreachable."
    echo "Run 'make forward-app' in a dedicated terminal window first."
    exit 1
end

echo "=== Starting load generator ($total_requests requests) ==="

for i in (seq 1 $total_requests)
    set -l endpoint (random choice $endpoints)

    # Generate mock W3C traceparent header: 00-<trace_id>-<span_id>-01
    set -l trace_id (head -c 16 /dev/urandom | xxd -p)
    set -l span_id (head -c 8 /dev/urandom | xxd -p)
    set -l traceparent "00-$trace_id-$span_id-01"

    # Send request in background and print HTTP status code, duration, and target route
    curl -s -o /dev/null -w "%{http_code} %{time_total}s $endpoint\n" \
        -H "traceparent: $traceparent" \
        "$base_url$endpoint" &

    # Synchronize batch every 10 requests to throttle concurrency
    if test (math "$i % 10") -eq 0
        wait
        sleep 0.1
    end
end

wait
echo "=== Load generation complete ==="
