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

	defer connection.Close()

	//store the incoming request in bytes
	bytes := make([]byte, 1024)
	n, errr := connection.Read(bytes)
	if errr != nil {
		fmt.Println("Error Reading request", errr)
		os.Exit(1)
	}

	msg := string(bytes[:n]) //convert the bytes to a string slice
	parts := strings.Split(msg, " ")
	request_target := parts[1];
	if len(parts) == 0 {
		fmt.Println("Error in retrieving the request target from the request")
		os.Exit(1)
	}
	if request_target == "/" {
		_, errs := connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if errs != nil {
			fmt.Println("Error in sending response back", errs)
		}
	} else {
		_, errs := connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"));
		if errs != nil {
			fmt.Println("Error in sending response back", errs)
		}
	}
}
