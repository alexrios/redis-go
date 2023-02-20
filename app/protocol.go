package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const NullBulkStr = "$-1\r\n"

type Command interface {
	Response() string
}

type PingCmd struct {
}

func (e PingCmd) Response() string {
	return fmt.Sprintf("+%s\r\n", "PONG")
}

type EchoCmd struct {
	params []string
}

func (e EchoCmd) Response() string {
	return fmt.Sprintf("+%s\r\n", strings.Join(e.params, " "))
}

type SetCmd struct {
	params     []string
	expiration int64
}

func (e SetCmd) Response() string {
	return fmt.Sprintf("+%s\r\n", "OK")
}

func (e SetCmd) Set(cache *Cache) error {
	if len(e.params) < 2 {
		return errors.New("no values to be set")
	}
	k := e.params[0]
	v := Value{
		val: e.params[1],
		exp: e.expiration,
	}
	err := cache.Store(k, v)
	if err != nil {
		return err
	}

	return nil
}

type GetCmd struct {
	params  []string
	value   string
	expired bool
}

func (e GetCmd) Response() string {
	if e.expired {
		return NullBulkStr
	}
	return fmt.Sprintf("+%s\r\n", e.value)
}

func (e *GetCmd) Get(cache *Cache) error {
	if len(e.params) < 1 {
		return errors.New("no values to be get")
	}
	k := e.params[0]
	v, err := cache.Load(k)
	if err != nil {
		return err
	}
	e.value = v
	return nil
}

var InvalidRequest = errors.New("invalid request")

func Decode(buffer []byte, cache *Cache) (Command, error) {
	tokens := bytes.Split(buffer, []byte("\r\n"))

	// tokens[0] is the header with a vaue like "*N", where N is the number of the tokens that follows
	header := tokens[0]
	// Fail when it's not an array.
	if header[0] != byte('*') {
		return nil, InvalidRequest
	}
	msgLen, err := strconv.Atoi(string(header[1])) // how much extra tokens should be parsed
	if err != nil {
		return nil, fmt.Errorf("%w: %s", InvalidRequest, err.Error())
	}

	// Create a new command with a given name
	cmd := string(tokens[2])

	paramsLen := msgLen - 1
	switch strings.ToUpper(cmd) {
	case "ECHO":
		return EchoCmd{params: ParseValues(paramsLen, tokens)}, nil
	case "PING":
		return PingCmd{}, nil
	case "SET":
		if paramsLen == 0 {
			return nil, errors.New("cannot SET with no parameters")
		}
		var exp int64 = 0
		values := ParseValues(paramsLen, tokens)
		for i := range values {
			if values[i] == "PX" && len(values) >= i+1 {
				millis, err := strconv.Atoi(values[i+1])
				if err != nil {
					return nil, err
				}
				exp = time.Now().UnixMilli() + int64(millis)
			}
		}
		setCmd := SetCmd{params: values, expiration: exp}
		err = setCmd.Set(cache)
		if err != nil {
			return nil, err
		}
		return setCmd, nil
	case "GET":
		if paramsLen == 0 {
			return nil, errors.New("cannot GET with no parameters")
		}
		values := ParseValues(paramsLen, tokens)
		getCmd := &GetCmd{params: values}

		err = getCmd.Get(cache)
		if err != nil {
			if errors.Is(err, KeyExpired) {
				getCmd.expired = true
				return getCmd, nil
			}
			return nil, err
		}
		return getCmd, nil
	}

	return nil, fmt.Errorf("no such command: %s", cmd)
}

func ParseValues(length int, tokens [][]byte) []string {
	values := make([]string, 0)

	paramsIndex := 4
	for i := 0; i < length; i++ {
		paramValuePos := paramsIndex + i*2
		values = append(values, string(tokens[paramValuePos]))
	}
	return values
}

func ErrorMsg(err error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err.Error()))
}
