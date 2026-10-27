package http1

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const maxLineLength = 4096
const maxHeaderCount = 100

// ParseRequest reads a HTTP-Request and parses it into the http.Request-struct from the
// standard go library. It returns an error if the request is malformed.
func ParseRequest(reader *bufio.Reader) (*http.Request, error) {

	request := &http.Request{}

	// Parse Line 1

	str1, err := readSingleLineFromReader(reader)
	if err != nil {
		return nil, err
	}
	str1param := strings.Split(str1, " ")

	if len(str1param) != 3 {
		return nil, fmt.Errorf("malformed request line: expected 3 parts, got %d", len(str1param))
	}

	reqUrl, err := url.ParseRequestURI(str1param[1])

	if err != nil {
		fmt.Printf("parser.go: failed to parse url")
		return nil, err
	}

	method := str1param[0]
	if !isValidMethod(method) {
		return nil, fmt.Errorf("invalid method: %q", method)
	}

	request.Method = str1param[0]
	request.URL = reqUrl
	request.Proto = str1param[2]

	// Parse Header

	request.Header = make(http.Header)

	currentHeaderCount := 0
	for {

		currentHeaderCount++
		if currentHeaderCount > maxHeaderCount {
			return nil, fmt.Errorf("too many header")
		}

		line, err := readSingleLineFromReader(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("headers truncated by EOF")
			}
			return nil, err
		}

		if len(line) == 0 {
			break
		}

		lines := strings.SplitN(string(line), ":", 2)

		if len(lines) != 2 {
			return nil, fmt.Errorf("malformed header line")
		}

		rawKey := strings.TrimSpace(lines[0])
		rawValue := strings.TrimSpace(lines[1])

		if !isValidHeaderKey(rawKey) {
			return nil, fmt.Errorf("invalid header key: %q", rawKey)
		}

		if strings.Contains(rawValue, "\t") {
			return nil, fmt.Errorf("invalid character in header value")
		}

		key := strings.TrimSpace(rawKey)
		value := strings.TrimSpace(rawValue)

		switch key {
		case "Host":
			request.Host = value
			continue
		}

		request.Header.Add(key, value)
	}

	te := request.Header.Get("Transfer-Encoding")
	if strings.EqualFold(strings.TrimSpace(te), "chunked") {
		data, err := parseChunkedBody(reader)
		if err != nil {
			return nil, err
		}
		request.ContentLength = int64(len(data))
		request.Body = io.NopCloser(bytes.NewReader(data))

	} else if clValues := request.Header.Values("Content-Length"); len(clValues) > 0 {
		firstVal := strings.TrimSpace(clValues[0])

		for _, v := range clValues[1:] {
			if strings.TrimSpace(v) != firstVal {
				return nil, fmt.Errorf("conflicting Content-Length values: %v", clValues)
			}
		}

		cl, err := strconv.ParseInt(firstVal, 10, 64)
		if err != nil || cl < 0 {
			return nil, fmt.Errorf("invalid Content-Length: %q", firstVal)
		}

		buf := make([]byte, cl)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, fmt.Errorf("reading Content-Length body: %w", err)
		}

		request.ContentLength = cl
		request.Body = io.NopCloser(bytes.NewReader(buf))
	} else {
		request.Body = http.NoBody
	}

	return request, nil
}

// isValidMethod checks if a method only contains valid characters
func isValidMethod(method string) bool {
	if len(method) == 0 {
		return false
	}

	for i := 0; i < len(method); i++ {
		c := method[i]
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'

		if !isUpper && !isLower {
			return false
		}
	}
	return true
}

// readSingleLineFromReader checks if a line ends correctly with \r\n and returns
// everything before it. error is nil on success and returns a string, otherwise
// error is not nil
func readSingleLineFromReader(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if len(line) > 0 {
			return "", fmt.Errorf("line truncated before CRNL")
		}
		return "", err
	}

	if len(line) > maxLineLength {
		return "", fmt.Errorf("line too long: %d bytes", len(line))
	}

	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("missing CR before NL")
	}

	line = line[:len(line)-2]

	if strings.Contains(line, "\r") {
		return "", fmt.Errorf("bare CR in line")
	}

	return line, nil
}

// isValidHeaderKey checks if all the condigtions for a key in a header are fulfulled
func isValidHeaderKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		isAlpha := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isDigit := c >= '0' && c <= '9'
		if !isAlpha && !isDigit && c != '-' {
			return false
		}
	}
	return true
}

// parseChunkedBody continues reading lines from the body until the chunk size is 0
// and returnes the data. error is nil on success.
func parseChunkedBody(reader *bufio.Reader) ([]byte, error) {
	var body bytes.Buffer

	for {
		line, err := readSingleLineFromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("reading chunk size: %v", err)
		}

		sizeStr := strings.TrimSpace(strings.Split(line, ";")[0])
		chunkSize, err := strconv.ParseInt(sizeStr, 16, 64)
		if err != nil || chunkSize < 0 {
			return nil, fmt.Errorf("invalid chunk size: %q", sizeStr)
		}

		if chunkSize == 0 {
			for {
				trailer, err := readSingleLineFromReader(reader)
				if err != nil {
					return nil, fmt.Errorf("reading trailer: %v", err)
				}
				if len(trailer) == 0 {
					break
				}
			}
			break
		}

		chunkData := make([]byte, chunkSize)
		if _, err := io.ReadFull(reader, chunkData); err != nil {
			return nil, fmt.Errorf("truncated chunk data: %v", err)
		}
		body.Write(chunkData)

		crnl := make([]byte, 2)
		if _, err := io.ReadFull(reader, crnl); err != nil {
			return nil, fmt.Errorf("reading chunk CRNL: %v", err)
		}
		if crnl[0] != '\r' || crnl[1] != '\n' {
			return nil, fmt.Errorf("malformed chunk: missing CRNL")
		}
	}

	return body.Bytes(), nil
}
