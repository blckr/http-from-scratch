package http1

import (
	"fmt"
	"net"
	"net/http"
)

// Response is a custom struct to handle a http-request
type Response struct {
	conn          net.Conn
	header        http.Header
	statusCode    int
	headerWritten bool
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
	if !r.headerWritten {
		r.WriteHeader(http.StatusOK)
	}
	return r.conn.Write(data)
}

// WriteHeader takes a status code and writes a corresponding header in the response
func (r *Response) WriteHeader(statusCode int) {
	if r.headerWritten {
		return
	}
	r.headerWritten = true
	r.statusCode = statusCode

	statusText := http.StatusText(statusCode)

	_, err := fmt.Fprintf(r.conn, "HTTP/1.1 %d %s\r\n", statusCode, statusText)
	if err != nil {
		return
	}
	err = r.header.Write(r.conn)
	if err != nil {
		return
	}

	fmt.Fprintf(r.conn, "\r\n")
}
