package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
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
	request_target, statusLine, body := getRequestTargetStatusLine(string(bytes))

	if request_target == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(request_target, "/echo") {
		val := strings.Split(request_target, "/")[2]
		encoding_format := getContentEncodingScheme(string(bytes))
		client_encodings := strings.Split(encoding_format, ",")
		for _, algo := range client_encodings {
			if strings.TrimSpace(algo) == "gzip" {
				connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: %s\r\n\r\n%s", len(val), algo, val)))
			}
		}
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)))

	} else if request_target == "/user-agent" {
		res := getUserAgent(string(bytes))
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(res), res)))
	} else if strings.HasPrefix(request_target, "/files/") {
		dir := os.Args[2]
		file_name := strings.Split(request_target, "/")[2]
		fmt.Println(dir + file_name)

		var method string = strings.Split(statusLine, " ")[0]
		if method == "GET" {
			file_content, err := os.ReadFile(dir + file_name)
			if err != nil {
				fmt.Println("Error in reding file: ", err)
				connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}
			connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file_content), string(file_content))))
		} else if method == "POST" {
			file, err := os.OpenFile(dir+file_name, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Println("error in opening file: ", err)
				os.Exit(1)
			}
			defer file.Close()
			length := getContentLength(string(bytes))
			fmt.Println("Before body: ", body)
			body := body[:length]
			fmt.Println("Body: ", body)
			_, err = file.WriteString(body)
			if err != nil {
				fmt.Println("Error in writing to the file: ", err)
				os.Exit(1)
			}
			connection.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	} else {
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func getRequestTargetStatusLine(request string) (string, string, string) {
	statusLine := strings.Split(request, "\r\n")[0]
	req_tar := strings.Split(statusLine, " ")[1]
	body := strings.Split(request, "\r\n\r\n")[1]
	fmt.Println(body)
	return req_tar, statusLine, body
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
func getContentLength(req string) int {
	headers := strings.Split(req, "\r\n")[1:]
	for _, v := range headers {
		if strings.HasPrefix(v, "Content-Length") {
			agent_value := v[16:]
			num, err := strconv.Atoi(agent_value)
			if err != nil {
				fmt.Println("Conversion error:", err)
				return 0
			}
			return num
		}
	}
	return 0
}

func getContentEncodingScheme(req string) string {
	headers := strings.Split(req, "\r\n")[1:]
	for _, v := range headers {
		if strings.HasPrefix(v, "Accept-Encoding") {
			agent_value := v[17:]
			return agent_value
		}
	}
	return ""
}
