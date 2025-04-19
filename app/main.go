package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	status_line    string
	headers        map[string]string
	body           string
	request_target string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4000")
	if err != nil {
		fmt.Println("Failed to bind to port 4000", err.Error())
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
	request_schema := extractRequestComponent(string(bytes))
	if request_schema.request_target == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(request_schema.request_target, "/echo") {
		val := strings.Split(request_schema.request_target, "/")[2]
		encoding_format := getContentEncodingScheme(request_schema.headers)
		client_encodings := strings.Split(encoding_format, ",")
		fmt.Println(client_encodings)
		for _, algo := range client_encodings {
			if strings.TrimSpace(algo) == "gzip" {
				compressed_data, err := compressData([]byte(val))
				if err != nil {
					break
				}
				connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: %s\r\n\r\n%s", len(compressed_data), algo, compressed_data)))
			}
		}
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)))

	} else if request_schema.request_target == "/user-agent" {
		res := getUserAgent(request_schema.headers)
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(res), res)))
	} else if strings.HasPrefix(request_schema.request_target, "/files/") {
		dir := os.Args[2]
		file_name := strings.Split(request_schema.request_target, "/")[2]
		fmt.Println(dir + file_name)

		var method string = strings.Split(request_schema.status_line, " ")[0]
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
			length := getContentLength(request_schema.headers)
			fmt.Println("Before body: ", request_schema.body)
			body := request_schema.body[:length]
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
func extractRequestComponent(request string) Request {
	status_line := strings.Split(request, "\r\n")[0]
	request_target := strings.Split(status_line, " ")[1]
	body := strings.Split(request, "\r\n\r\n")[1]
	header_payload := strings.Split(request, "\r\n")[1:]
	header_payload = header_payload[:len(header_payload)-1]
	headers := make(map[string]string)
	for _, v := range header_payload {
		if strings.Contains(v, ":") {
			header := strings.Split(v, ": ")
			headers[header[0]] = header[1]
		}
	}
	req_tar := Request{
		status_line:    status_line,
		headers:        headers,
		body:           body,
		request_target: request_target,
	}
	return req_tar
}

func getUserAgent(headers map[string]string) string {
	if val, ok := headers["User-Agent"]; ok {
		return strings.TrimSpace(val)
	}
	return ""
}
func getContentLength(headers map[string]string) int {
	if val, ok := headers["Content-Length"]; ok {
		length, err := strconv.Atoi(val)
		if err != nil {
			fmt.Println("Error in converting content length to int: ", err)
			return 0
		}
		return length
	}
	return 0
}

func getContentEncodingScheme(headers map[string]string) string {
	if val, ok := headers["Accept-Encoding"]; ok {
		return strings.TrimSpace(val)
	}
	return "";
}

func compressData(d []byte) (string, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(d)
	if err != nil {
		return "", err
	}
	writer.Close()
	return buf.String(), nil
}
