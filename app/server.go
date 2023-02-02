package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
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
		fmt.Println("send to conn channel")
		go func() {
			fmt.Println("new routine")
			for {

				// store incoming data
				buffer := make([]byte, 1024)
				_, err = conn.Read(buffer)
				if err != nil {
					fmt.Println(err)
					// notify the program that a connection is not available anymore.
					conn.Close()
					return
				}
				switch {
				case bytes.Contains(buffer[1:], []byte("PING")):
					fmt.Println("responding pong")
					_, err = conn.Write([]byte("+PONG\r\n"))
				case bytes.Contains(buffer, []byte("DOCS")):
					fmt.Println("responding docs")
					_, err = conn.Write([]byte("+welcome to redis\r\n"))
				}
				if err != nil {
					conn.Close()
					fmt.Println("Error: ", err.Error())
					os.Exit(1)
				}
			}
		}()
	}
}
