package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

type Reponse struct {
	msg  []byte
	conn net.Conn
}

func (r Reponse) isStringCmd() bool {
	return bytes.HasPrefix(r.msg, []byte("+"))
}

func (r Reponse) StrMsg() string {
	return string(r.msg)
}

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

	conns := make(chan net.Conn)
	deadConns := make(chan net.Conn)
	responses := make(chan Reponse)

	// send to the connections channel.
	go func() {
		for {
			var conn net.Conn
			conn, err := l.Accept()
			fmt.Println("send to conn channel")
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
			fmt.Println("conns")
			go func() {
				fmt.Println("new routine")
				// store incoming data
				buffer := make([]byte, 1024)
				_, err = conn.Read(buffer)
				if err != nil {
					fmt.Println(err)
					// notify the program that a connection is not available anymore.
					fmt.Println("dead from read")
					deadConns <- conn
					return
				}
				fmt.Println("send to responses channel")
				responses <- Reponse{
					msg:  buffer,
					conn: conn,
				}
			}()
		case dc := <-deadConns:
			fmt.Println("closing conn")
			_ = dc.Close()
		case r := <-responses:
			// handle protocol messages
			var wError error
			switch {
			case bytes.Contains(r.msg[1:], []byte("PING")):
				fmt.Println("responding pong")
				_, wError = r.conn.Write([]byte("+PONG\r\n"))
				//default:
				//	fmt.Println("responding error")
				//	_, wError = r.conn.Write([]byte("-invalid data type\r\n"))
			}
			if wError != nil {
				fmt.Println("dead from write")
				fmt.Println(wError)
				// notify the program that a connection is not available anymore.
				deadConns <- r.conn
			}
			r.conn.Close()

		}
	}
}
