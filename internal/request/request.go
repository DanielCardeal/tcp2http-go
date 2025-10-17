package request

import (
	"bytes"
	"fmt"
	"io"
)

type ParserState int

const (
	PARSER_STATE_INITIALIZED ParserState = iota
	PARSER_STATE_DONE
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

	fields := bytes.Split(requestLine, []byte(" "))
	if len(fields) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	versionParts := bytes.Split(fields[2], []byte("/"))
	if len(versionParts) != 2 ||
		!bytes.Equal(versionParts[0], []byte("HTTP")) ||
		!bytes.Equal(versionParts[1], []byte("1.1")) {
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

	return r, idx, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case PARSER_STATE_INITIALIZED:
		rl, n, err := parseRequestLine(data)
		if err != nil || n == 0 {
			return 0, err
		}
		r.RequestLine = *rl
		r.state = PARSER_STATE_DONE
		return n, err
	case PARSER_STATE_DONE:
		return 0, nil
	default:
		return 0, ERROR_UNKNOWN_PARSER_STATE
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var request Request
	data := make([]byte, 0, 4096)
	readBuff := make([]byte, 4096)
	for request.state != PARSER_STATE_DONE {
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
