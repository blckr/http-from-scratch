package http1

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ParseRequest reads a HTTP-Request and parses it into the http.Request-struct from the
// standard go library. It returns an error if the request is malformed.
func ParseRequest(reader *bufio.Reader) (*http.Request, error) {

	request := &http.Request{}

	// Parse Line 1

	line1, isPrefix, err := reader.ReadLine()

	if err != nil {
		return nil, err
	}
	if isPrefix {
		return nil, fmt.Errorf("request line too long")
	}

	str1 := string(line1)

	str1param := strings.Split(str1, " ")

	if len(str1param) != 3 {
		return nil, fmt.Errorf("malformed request line: expected 3 parts, got %d", len(str1param))
	}

	reqUrl, err := url.ParseRequestURI(str1param[1])

	if err != nil {
		fmt.Printf("parser.go: failed to parse url")
		return nil, err
	}

	request.Method = str1param[0]
	request.URL = reqUrl
	request.Proto = str1param[2]

	// Parse Header

	request.Header = make(http.Header)

	for {
		line, isPrefix, err := reader.ReadLine()

		if isPrefix || err != nil {
			fmt.Println("parser.go: failed to read header-line")
			return nil, err
		}

		if len(line) == 0 {
			break
		}

		lines := strings.SplitN(string(line), ":", 2)

		if len(lines) != 2 {
			return nil, fmt.Errorf("malformed header line")
		}

		key := strings.TrimSpace(lines[0])
		value := strings.TrimSpace(lines[1])

		switch key {
		case "Host":
			request.Host = value
			continue
		}

		request.Header.Add(key, value)
	}

	return request, nil
}
