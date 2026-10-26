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

func TestParseRequest_TwoHeader(t *testing.T) {

	rawInput := "GET /hello HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Custom: Header\r\n" +
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

	if req.Header.Get("Custom") != "Header" {
		t.Errorf("expected header 'Header', got: %q", req.Host)
	}
}

func TestParseRequest_IntroExample(t *testing.T) {

	rawInput := "GET /index.html HTTP/1.1\r\n" +
		"Host: www.example.com\r\n" +
		"User-Agent: Browser\r\n" +
		"\r\n"

	reader := bufio.NewReader(strings.NewReader(rawInput))

	req, err := ParseRequest(reader)
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("expected method 'GET', got: %q", req.Method)
	}

	if req.URL == nil || req.URL.Path != "/index.html" {
		t.Errorf("expected path '/index.html', got: %v", req.URL)
	}

	if req.Host != "www.example.com" {
		t.Errorf("expected host 'www.example.com', got: %q", req.Host)
	}

	if req.Header.Get("User-Agent") != "Browser" {
		t.Errorf("expected header 'Browser', got: %q", req.Host)
	}
}
