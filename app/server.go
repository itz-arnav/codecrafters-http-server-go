package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func handlePostFile(connection net.Conn, inputHeaderStringList []string, directoryPath string) error {
	firstLineOfString := inputHeaderStringList[0]
	firstLineParts := strings.Split(firstLineOfString, " ")
	fileName := firstLineParts[1][7:]

	resultFileContent := strings.Join(inputHeaderStringList[6:], " ")
	resultFileContent = strings.ReplaceAll(resultFileContent, "\x00", "")
	resultFileContent = strings.TrimSuffix(resultFileContent, "\n")
	resultFileContent = strings.TrimSuffix(resultFileContent, "\r")

	file, err := os.Create(directoryPath + "/" + fileName)
	if err != nil {
		return fmt.Errorf("error while creating file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(resultFileContent); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	responseHeaders := "HTTP/1.1 201 OK\r\n\r\nContent-Type: text/plain\r\n\r\n"
	if _, err := connection.Write([]byte(responseHeaders)); err != nil {
		return fmt.Errorf("error writing HTTP headers: %w", err)
	}
	return nil
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	buf := make([]byte, 1024)
	n, err := connection.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while reading from connection: %s", err)
		return
	}
	myString := string(buf[:n])

	inputHeaderStringList := strings.Split(myString, "\r\n")
	if len(inputHeaderStringList) < 1 {
		fmt.Println("Error: received empty request")
		return
	}

	firstLineOfString := inputHeaderStringList[0]
	firstLineParts := strings.Split(firstLineOfString, " ")
	if len(firstLineParts) != 3 {
		fmt.Println("Error parsing request line: ", firstLineOfString)
		return
	}

	// Handle different request types
	switch {
	case strings.Contains(firstLineParts[1], "/echo/"):
		handleEchoRequest(connection, firstLineParts)
	case firstLineParts[1] == "/user-agent":
		handleUserAgentRequest(connection, inputHeaderStringList)
	case strings.Contains(firstLineParts[1], "/files/"):
		handleFileRequest(connection, firstLineParts, inputHeaderStringList)
	case firstLineParts[1] == "/":
		sendResponse(connection, "HTTP/1.1 200 OK\r\n\r\nContent-Type: text/plain\r\n\r\n")
	default:
		sendResponse(connection, "HTTP/1.1 404 Not Found\r\n\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n")
	}
}

func handleEchoRequest(connection net.Conn, firstLineParts []string) {
	contentString := firstLineParts[1][6:]
	sendTextResponse(connection, contentString, 200)
}

func handleUserAgentRequest(connection net.Conn, inputHeaderStringList []string) {
	contentString := strings.Split(inputHeaderStringList[2], ": ")[1]
	sendTextResponse(connection, contentString, 200)
}

func handleFileRequest(connection net.Conn, firstLineParts []string, inputHeaderStringList []string) {
	if len(os.Args) != 3 {
		fmt.Println("Invalid command-line arguments provided.")
		return
	}

	directoryPath := os.Args[2]
	if firstLineParts[0] == "POST" {
		if err := handlePostFile(connection, inputHeaderStringList, directoryPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to handle POST file: %s", err)
			sendResponse(connection, "HTTP/1.1 500 Internal Server Error\r\n\r\nContent-Type: text/plain\r\n\r\n")
		}
		return
	}

	// Handle GET file request
	fileName := firstLineParts[1][7:]
	filePath := directoryPath + "/" + fileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		sendResponse(connection, "HTTP/1.1 404 Not Found\r\n\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n")
		return
	}

	sendFileResponse(connection, filePath)
}

func sendFileResponse(connection net.Conn, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %s", err)
		sendResponse(connection, "HTTP/1.1 500 Internal Server Error\r\n\r\nContent-Type: text/plain\r\n\r\n")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %s", err)
		sendResponse(connection, "HTTP/1.1 500 Internal Server Error\r\n\r\nContent-Type: text/plain\r\n\r\n")
		return
	}

	contentString := string(content)
	contentLength := strconv.Itoa(len(contentString))
	responseHeaders := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %s\r\n\r\n%s", contentLength, contentString)
	sendResponse(connection, responseHeaders)
}

func sendTextResponse(connection net.Conn, contentString string, statusCode int) {
	contentLength := strconv.Itoa(len(contentString))
	responseHeaders := fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: text/plain\r\nContent-Length: %s\r\n\r\n%s", statusCode, contentLength, contentString)
	sendResponse(connection, responseHeaders)
}

func sendResponse(connection net.Conn, responseHeaders string) {
	if _, err := connection.Write([]byte(responseHeaders)); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending response: %s", err)
	}
}

func main() {
	fmt.Println("Server is starting...")

	server, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind to port 4221: %s", err)
		os.Exit(1)
	}
	defer server.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %s", err)
			continue
		}

		go handleConnection(connection)
	}
}
