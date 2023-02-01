package main

import (
	"fmt"
	"net"
	"os"
)

type Reponse struct {
	msg  []byte
	conn net.Conn
}

func (r Reponse) StrMsg() string {
	return string(r.msg)
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer func(l net.Listener) {
		if err := l.Close(); err != nil {
			fmt.Println("Failed to close the tcp listener: %w", err)
			os.Exit(1)
		}
	}(l)

	conns := make(chan net.Conn)
	deadConns := make(chan net.Conn)
	responses := make(chan Reponse)

	// send to the connections channel.
	go func() {
		for {
			var conn net.Conn
			conn, err := l.Accept()
			conns <- conn

			if err != nil {
				fmt.Println("Error accepting connection: ", err.Error())
				os.Exit(1)
			}
		}
	}()

	// Iterate over all channels to decide what to do.
	// THIS IS THE EVENT LOOP
	for {
		select {
		// brand new connecion
		case conn := <-conns:
			// new routine for a new connection
			go func() {
				// store incoming data
				buffer := make([]byte, 1024)
				_, err = conn.Read(buffer)
				if err != nil {
					fmt.Println(err)
					// notify the program that a connection is not available anymore.
					deadConns <- conn
					return
				}
				responses <- Reponse{
					msg:  buffer,
					conn: conn,
				}
			}()
		case dc := <-deadConns:
			_ = dc.Close()
		case r := <-responses:
			// handle protocol messages
			var wError error
			switch r.StrMsg() {
			case "PING":
				_, wError = r.conn.Write([]byte("+PONG\r\n"))
			default:
				_, wError = r.conn.Write([]byte("+WTF\r\n"))
			}
			if wError != nil {
				fmt.Println(wError)
				// notify the program that a connection is not available anymore.
				deadConns <- r.conn
			}

		}
	}
}
