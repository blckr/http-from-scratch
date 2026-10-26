package http1

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseRequest_SimpleGet(t *testing.T) {

	rawInput := "GET /hello HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"\r\n"

	reader := bufio.NewReader(strings.NewReader(rawInput))

	req, err := ParseRequest(reader)
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("expected method 'GET', got: %q", req.Method)
	}

	if req.URL == nil || req.URL.Path != "/hello" {
		t.Errorf("expected path '/hello', got: %v", req.URL)
	}

	if req.Host != "localhost:8080" {
		t.Errorf("expected host 'localhost:8080', got: %q", req.Host)
	}
}
