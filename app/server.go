package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	server, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer server.Close()

	connection, err := server.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer connection.Close()

	buf := make([]byte, 1024)
	_, err = connection.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while making buffer: %s", err)
		os.Exit(1)
	}

	myString := string(buf)
	index := strings.Index(myString, "\r\n")
	firstLineOfString := myString[:index]
	firstLineParts := strings.Split(firstLineOfString, " ")

	if len(firstLineParts) != 3 {
		fmt.Println("Error writing parsing string: ", firstLineOfString)
		os.Exit(1)
	}

	if firstLineParts[1] == "/" {
		responseHeaders := "HTTP/1.1 200 OK\r\n\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"
		_, err = connection.Write([]byte(responseHeaders))
		if err != nil {
			fmt.Println("Error writing HTTP header: ", err.Error())
			os.Exit(1)
		}
	} else {
		responseHeaders := "HTTP/1.1 404 Not Found\r\n\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"
		_, err = connection.Write([]byte(responseHeaders))
		if err != nil {
			fmt.Println("Error writing HTTP header: ", err.Error())
			os.Exit(1)
		}
	}

}
