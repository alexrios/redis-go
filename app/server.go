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
	"sync"
)

var _map = sync.Map{}

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

	CommandSwitch:
		switch strings.ToUpper(cmd) {
		case "ECHO":
			resp := fmt.Sprintf("+%s\r\n", strings.Join(values, " "))
			_, err = conn.Write([]byte(resp))
		case "PING":
			_, err = conn.Write([]byte("+PONG\r\n"))
		case "SET":
			err = setVar(values)
			if err != nil {
				break CommandSwitch
			}
			_, err = conn.Write([]byte("+OK\r\n"))
		case "GET":
			v, err := getVar(values)
			if err != nil {
				break CommandSwitch
			}
			_, err = conn.Write([]byte(fmt.Sprintf("+%s\r\n", v)))
		default:
			_, err = conn.Write([]byte(fmt.Sprintf("-%s is not a valid command\r\n", cmd)))
		}

		if err != nil {
			_, err = conn.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
		}
	}
}

func getVar(values []string) (string, error) {
	if len(values) < 1 {
		return "", errors.New("no values to be get")
	}
	rawV, ok := _map.Load(values[0])
	if !ok {
		return "", errors.New("key not found")
	}
	v, ok := rawV.(string)
	if !ok {
		return "", errors.New("key not a string")
	}
	return v, nil
}

func setVar(values []string) error {
	if len(values) < 2 {
		return errors.New("no values to be set")
	}

	_map.Store(values[0], values[1])

	return nil
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
		if paramsLen == 0 {
			return cmd, nil
		}
		paramsIndex := 4
		for i := 0; i < paramsLen; i++ {
			paramValuePos := paramsIndex + i*2
			values = append(values, string(tokens[paramValuePos]))
		}
		return cmd, values
	}
	return "", nil
}
