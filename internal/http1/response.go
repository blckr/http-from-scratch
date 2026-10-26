package http1

import (
	"fmt"
	"net"
	"net/http"
)

type Response struct {
	conn          net.Conn
	header        http.Header
	statusCode    int
	headerWritten bool
}

func NewResponse(conn net.Conn) *Response {
	header := make(http.Header)
	return &Response{conn: conn, header: header, statusCode: 200, headerWritten: false}
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Write(data []byte) (n int, err error) {
	if !r.headerWritten {
		r.WriteHeader(http.StatusOK)
	}
	return r.conn.Write(data)
}

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
