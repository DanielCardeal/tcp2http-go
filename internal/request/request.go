package request

import (
	"bytes"
	"fmt"
	"io"
)

type ParserState int

const (
	StateInitialized ParserState = iota
	StateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       ParserState
}

var SEPARATOR = []byte("\r\n")

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
var ERROR_INVALID_HTTP_METHOD = fmt.Errorf("unknown HTTP method")
var ERROR_UNKNOWN_PARSER_STATE = fmt.Errorf("unknown parser state")

func (r *RequestLine) ValidMethod() bool {
	return r.Method == "GET" || r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE"
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	requestLine := b[:idx]
	read := idx + len(SEPARATOR)

	fields := bytes.Split(requestLine, []byte(" "))
	if len(fields) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	versionParts := bytes.Split(fields[2], []byte("/"))
	if len(versionParts) != 2 ||
		string(versionParts[0]) != "HTTP" ||
		string(versionParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	r := &RequestLine{
		Method:        string(fields[0]),
		RequestTarget: string(fields[1]),
		HttpVersion:   string(versionParts[1]),
	}
	if !r.ValidMethod() {
		return nil, 0, ERROR_INVALID_HTTP_METHOD
	}

	return r, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case StateInitialized:
		rl, n, err := parseRequestLine(data)
		if err != nil || n == 0 {
			return 0, err
		}
		r.RequestLine = *rl
		r.state = StateDone
		return n, err
	case StateDone:
		return 0, nil
	default:
		return 0, ERROR_UNKNOWN_PARSER_STATE
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var request Request
	data := make([]byte, 0, 4096)
	readBuff := make([]byte, 4096)
	for request.state != StateDone {
		n, err := reader.Read(readBuff)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		data = append(data, readBuff[:n]...)

		n, err = request.parse(data)
		if err != nil {
			return nil, err
		}
		if n != 0 {
			newData := make([]byte, 0, len(data[n:]))
			copy(newData, data[n:])
			data = newData
		}
	}

	return &request, nil
}
