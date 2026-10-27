package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
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
				sendError(conn, http.StatusBadRequest)
				return nil
			}
		}

		if ok, statusCode := checkHTTPVersion(request.Proto); !ok {
			sendError(conn, statusCode)
			return nil
		}

		response := NewResponse(conn)

		mux.ServeHTTP(response, request)

		if err := response.Flush(); err != nil {
			return err
		}

		connHdr := request.Header.Get("Connection")

		if strings.EqualFold(connHdr, "close") {
			break
		}

		if request.Proto == "HTTP/1.0" && !strings.EqualFold(connHdr, "keep-alive") {
			break
		}

	}
	return nil

}

func sendError(conn net.Conn, statusCode int) {

	statusText := http.StatusText(statusCode)

	if statusText == "" {
		statusText = "Error"
	}

	body := fmt.Sprintf("%d %s\n", statusCode, statusText)

	resp := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s",
		statusCode, statusText, len(body), body,
	)

	_, _ = conn.Write([]byte(resp))
}

func checkHTTPVersion(version string) (bool, int) {

	if !strings.HasPrefix(version, "HTTP/") {
		return false, 400
	}

	versionNumber := strings.TrimPrefix(version, "HTTP/")

	parts := strings.Split(versionNumber, ".")
	if len(parts) != 2 {
		return false, 400
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, 400
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, 400
	}

	allowedVersions := map[[2]int]bool{
		{1, 0}: true,
		{1, 1}: true,
	}

	if allowedVersions[[2]int{major, minor}] {
		return true, 0
	} else {
		return false, 505
	}

}
