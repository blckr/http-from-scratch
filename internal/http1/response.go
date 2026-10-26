package http1

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// Response is a custom struct to handle a http-request
type Response struct {
	conn          net.Conn
	header        http.Header
	statusCode    int
	headerWritten bool

	body bytes.Buffer
}

// NewResponse takes a connection and returns a new Response with default values
func NewResponse(conn net.Conn) *Response {
	header := make(http.Header)
	return &Response{conn: conn, header: header, statusCode: 200, headerWritten: false}
}

// Header returns the header in the passed response
func (r *Response) Header() http.Header {
	return r.header
}

// Write adds data to the response
func (r *Response) Write(data []byte) (n int, err error) {
	return r.body.Write(data)
}

// WriteHeader takes a status code and writes a corresponding header in the response
func (r *Response) WriteHeader(statusCode int) {
	if r.headerWritten {
		return
	}
	r.statusCode = statusCode
	r.headerWritten = true
}

func (r *Response) Flush() error {

	if r.header.Get("Content-Length") == "" {
		r.header.Set("Content-Length", strconv.Itoa(r.body.Len()))
	}

	statusText := http.StatusText(r.statusCode)
	if statusText == "" {
		statusText = "Unknown"
	}

	if _, err := fmt.Fprintf(r.conn, "HTTP/1.1 %d %s\r\n", r.statusCode, statusText); err != nil {
		return err
	}

	if err := r.header.Write(r.conn); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.conn, "\r\n"); err != nil {
		return err
	}

	_, err := r.conn.Write(r.body.Bytes())
	return err
}
