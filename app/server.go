package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

func main() {
	var _map = Cache{
		v: make(map[string]Value, 0),
		m: sync.RWMutex{},
	}

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
		go handleConnection(conn, &_map)
	}
}

func handleConnection(conn net.Conn, cache *Cache) {
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
		cmd, err := Decode(buffer, cache)
		if err != nil {
			_, err = conn.Write(ErrorMsg(err))
			continue
		}
		_, err = conn.Write([]byte(cmd.Response()))
		if err != nil {
			_, err = conn.Write(ErrorMsg(err))
		}
	}
}
