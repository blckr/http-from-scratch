package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	for {

		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go HandleConnection(conn, handler)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {

		request, err := ParseRequest(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		}

		response := NewResponse(conn)

		mux.ServeHTTP(response, request)

		if err := response.Flush(); err != nil {
			return err
		}

		if strings.EqualFold(request.Header.Get("Connection"), "close") {
			break
		}

	}
	return nil

}
