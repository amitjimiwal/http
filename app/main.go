package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer connection.Close() //closing the connection after main functions exits with success, error or panic , anything

	//store the incoming request in bytes
	bytes := make([]byte, 1024)
	n, errr := connection.Read(bytes)
	if errr != nil {
		fmt.Println("Error Reading request", errr)
		os.Exit(1)
	}

	msg := string(bytes[:n]) //convert the bytes to a string slice
	parts := strings.Split(msg, " ")
	request_target := parts[1]
	if len(parts) == 0 {
		fmt.Println("Error in retrieving the request target from the request")
		os.Exit(1)
	}
	//extra check for extracting the str from /echo/{str}
	if !strings.HasPrefix(request_target, "/echo/") {
		fmt.Println("The requested resource is not present")
		os.Exit(1)
	}
	//extract value from request target
	value := strings.Split(request_target, "/")[2];
	connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(value), value)))
}
