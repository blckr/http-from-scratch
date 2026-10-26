package http1

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {

	request := &http.Request{}

	// Parse Line 1

	line1, isPrefix, err := reader.ReadLine()

	if isPrefix == true || err != nil {
		fmt.Println("parser.go: reading line1 failed")
	}

	str1 := string(line1)

	str1param := strings.Split(str1, " ")

	if len(str1param) != 3 {
		fmt.Println("parser.go: str1param is too long")
	}

	reqUrl, err := url.ParseRequestURI(str1param[1])

	if err != nil {
		fmt.Printf("parser.go: failed to parse url")
	}

	request.Method = str1param[0]
	request.URL = reqUrl
	request.Proto = str1param[2]

	// Parse Line 2

	line2, isPrefix, err := reader.ReadLine()

	if isPrefix == true || err != nil {
		fmt.Println("parser.go: reading line1 failed")
	}

	str2 := string(line2)

	str2param := strings.Split(str2, " ")

	if len(str2param) != 2 {
		fmt.Printf("parser.go: str2param has an unexpected length")
	}

	request.Host = str2param[1]

	// Loop

	// TODO

	// Done

	return request, nil
}
