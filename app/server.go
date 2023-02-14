package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Listen on port
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("lintening")
	// Unbind port before the program exits
	defer func(l net.Listener) {
		if err := l.Close(); err != nil {
			fmt.Println("Failed to close the tcp listener: %w", err)
			os.Exit(1)
		}
	}(l)

	for {

		var conn net.Conn
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println(err)
			os.Exit(1)
		}

		// commands
		cmd, values := Decode(buffer)

		if cmd == "ECHO" {
			resp := fmt.Sprintf("+%s\r\n", strings.Join(values, " "))
			_, err = conn.Write([]byte(resp))
			if err != nil {
				fmt.Println("Error: ", err.Error())
			}
			return
		}

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error: ", err.Error())
		}
	}
}

func Decode(buffer []byte) (string, []string) {
	tokens := bytes.Split(buffer, []byte("\r\n"))

	header := tokens[0]
	typpe := header[0]
	typeLen, err := strconv.Atoi(string(header[1]))
	if err != nil {
		fmt.Println(err)
	}
	cmd := string(tokens[2])
	if typpe == byte('*') {
		values := make([]string, 0)
		paramsLen := typeLen - 1
		for i := 0; i <= paramsLen; i = i + 2 {
			values = append(values, string(tokens[4+i]))
		}
		return cmd, values
	}
	return "", nil
}
