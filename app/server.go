package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func handleConnection(connection net.Conn) {
	defer connection.Close()

	buf := make([]byte, 1024)
	_, err := connection.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while making buffer: %s", err)
		os.Exit(1)
	}

	myString := string(buf)
	inputHeaderStringList := strings.Split(myString, "\r\n")
	firstLineOfString := inputHeaderStringList[0]
	firstLineParts := strings.Split(firstLineOfString, " ")

	if len(firstLineParts) != 3 {
		fmt.Println("Error writing parsing string: ", firstLineOfString)
		os.Exit(1)
	}

	if strings.Contains(firstLineParts[1], "/echo/") {
		contentString := firstLineParts[1][6:]
		contentLength := strconv.Itoa(len(contentString))
		responseHeaders := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" + "Content-Length: " + contentLength + "\r\n\r\n" + contentString + "\r\n"

		_, err = connection.Write([]byte(responseHeaders))
		if err != nil {
			fmt.Println("Error writing HTTP header: ", err.Error())
			os.Exit(1)
		}
	} else if firstLineParts[1] == "/user-agent" {
		contentString := strings.Split(inputHeaderStringList[2], ": ")[1]
		contentLength := strconv.Itoa(len(contentString))
		responseHeaders := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" + "Content-Length: " + contentLength + "\r\n\r\n" + contentString + "\r\n"
		_, err = connection.Write([]byte(responseHeaders))
		if err != nil {
			fmt.Println("Error writing HTTP header: ", err.Error())
			os.Exit(1)
		}
	} else if strings.Contains(firstLineParts[1], "/files/") {

		if len(os.Args) != 3 {
			fmt.Println("Invalid command-line arguments provided.")
			os.Exit(1)
		}

		directoryPath := os.Args[2]
		files, err := os.ReadDir(directoryPath)
		if err != nil {
			fmt.Println("Invalid permissions to read the directory.")
			os.Exit(1)
		}

		fileName := firstLineParts[1][7:]
		for _, file := range files {
			if file.Name() == fileName {
				filePath := directoryPath + "/" + file.Name()
				file, err := os.Open(filePath)
				if err != nil {
					fmt.Println("Invalid permissions to open the file.")
					fmt.Println("File Path: ", filePath)
					os.Exit(1)
				}
				defer file.Close()

				content, err := io.ReadAll(file)
				if err != nil {
					fmt.Println("Failed to read the file")
					os.Exit(1)
				}
				contentString := string(content)
				responseHeaders := "HTTP/1.1 200 OK\r\n" +
					"Content-Type: application/octet-stream\r\n" + contentString + "\r\n\r\n"
				_, err = connection.Write([]byte(responseHeaders))
				if err != nil {
					fmt.Println("Error writing HTTP header: ", err.Error())
					os.Exit(1)
				}
			}
		}
		responseHeaders := "HTTP/1.1 404 Not Found\r\n\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"
		_, err = connection.Write([]byte(responseHeaders))
		if err != nil {
			fmt.Println("Error writing HTTP header: ", err.Error())
			os.Exit(1)
		}

	} else if firstLineParts[1] == "/" {
		responseHeaders := "HTTP/1.1 200 OK\r\n\r\n" +
			"Content-Type: text/plain\r\n\r\n"
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	server, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer server.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		// Handle each connection concurrently in a new goroutine
		go handleConnection(connection)
	}

}
