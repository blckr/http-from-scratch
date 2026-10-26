#!/usr/bin/env fish

go build -o srv ./cmd/httpserver
./srv &
set -l pid $last_pid

sleep 0.8

echo "--- GET /hello ---"
echo "http://localhost:8080/hello"
echo "---:"
curl -i http://localhost:8080/hello

echo -e "\n--- GET /unknown (404) ---"
echo "http://localhost:8080/unknown"
echo "---:"
curl -i http://localhost:8080/unknown

kill $pid
rm srv
