package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestBytes(t *testing.T) {
	buffer := []byte("*3\r\n$4\r\nECHO\r\n$3\r\nhey\r\n$3\r\nhey\r\n")
	//buffer := []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n")
	//buffer := []byte("*1\r\n$4\r\nECHO\r\n")

	//fmt.Printf("%q\n", bytes.Split(buffer, []byte("\r\n")))

	//tokens := bytes.Split(buffer, []byte("\r\n"))
	// the sequence is an array?
	if buffer[0] == byte('*') {
		startIdx := 1
		endIdx := startIdx + 1
		for ; buffer[endIdx] != '\r'; endIdx++ {
		}
		cmdLen, err := strconv.Atoi(string(buffer[startIdx:endIdx]))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("array len:", cmdLen)

		for ; buffer[startIdx] != '$'; startIdx++ {
		}
		//starting from first command
		endIdx = startIdx + 1
		for ; buffer[endIdx] != '\r'; endIdx++ {
		}

		endIdx++
		startIdx = endIdx + 1
		// find next \r
		for ; buffer[endIdx] != '\r'; endIdx++ {
		}

		cmd := string(buffer[startIdx:endIdx])
		// early return when no arguments passed
		if cmdLen == 0 {
			fmt.Println("CMD", cmd)
			fmt.Println("VALUES", nil)
			return
		}

		if cmd == "ECHO" {
			for endIdx < len(buffer) {
				// find next $
				for ; buffer[startIdx] != '$'; startIdx++ {
				}
				// find next \r
				for ; buffer[endIdx] != '\r'; endIdx++ {
				}
				fmt.Println(string(buffer[startIdx:endIdx]))
				endIdx++
			}
		}
	}
}

func TestStrings(t *testing.T) {
	buffer := []byte("*3\r\n$4\r\nECHO\r\n$3\r\nhey\r\n$3\r\nhey\r\n")
	//buffer := []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n")
	//buffer := []byte("*1\r\n$4\r\nECHO\r\n")

	Decode(buffer, nil)

}
