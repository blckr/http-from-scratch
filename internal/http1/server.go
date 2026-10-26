package http1

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
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

		/*
				Functions that might be useful:
				* listener.Accept()
			  * HandleConnection()
		*/
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, err := ParseRequest(reader)
	if err != nil {
		return err
	}

	response := NewResponse(conn)

	mux.ServeHTTP(response, request)
	return nil
}
