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
	counter := 1
	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Request No %d \n", counter)
		counter++
		go handleReq(connection) //goroutine
	}

}
func handleReq(connection net.Conn) {
	defer connection.Close() //closing the connection after main functions exits with success, error or panic , anything

	//store the incoming request in bytes
	bytes := make([]byte, 1024)
	_, errr := connection.Read(bytes)
	if errr != nil {
		fmt.Println("Error Reading request", errr)
		os.Exit(1)
	}
	request_target := getRequestTarget(string(bytes))

	if request_target == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(request_target, "/echo") {
		val := strings.Split(request_target, "/")[2]
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)))
	} else if request_target == "/user-agent" {
		res := getUserAgent(string(bytes))
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(res), res)))
	} else if strings.HasPrefix(request_target, "/files/") {
		dir := os.Args[2]
		file_name := strings.Split(request_target, "/")[2]
		fmt.Println(dir + file_name)
		file_content, err := os.ReadFile(dir + file_name)
		if err != nil {
			fmt.Println("Error in reding file: ", err)
			connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file_content), string(file_content))))
	} else {
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func getRequestTarget(request string) string {
	statusLine := strings.Split(request, "\r\n")[0]
	req_tar := strings.Split(statusLine, " ")[1]
	return req_tar
}

func getUserAgent(req string) string {
	headers := strings.Split(req, "\r\n")[1:]
	for _, v := range headers {
		if strings.HasPrefix(v, "User-Agent") {
			agent_value := v[11:]
			return strings.TrimSpace(agent_value)
		}
	}
	return ""
}
