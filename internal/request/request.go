package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var SEPARATOR = []byte("\r\n")

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
var ERROR_INVALID_HTTP_METHOD = fmt.Errorf("unknown HTTP method")

func (r *RequestLine) ValidMethod() bool {
	return r.Method == "GET" || r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE"
}

func parseRequestLine(b []byte) (*RequestLine, []byte, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, b, nil
	}
	requestLine := b[:idx]
	restOfMsg := b[idx+len(SEPARATOR):]

	fields := bytes.Split(requestLine, []byte(" "))
	if len(fields) != 3 {
		return nil, b, ERROR_MALFORMED_REQUEST_LINE
	}

	versionParts := bytes.Split(fields[2], []byte("/"))
	if len(versionParts) != 2 ||
		!bytes.Equal(versionParts[0], []byte("HTTP")) ||
		!bytes.Equal(versionParts[1], []byte("1.1")) {
		return nil, b, ERROR_MALFORMED_REQUEST_LINE
	}

	r := &RequestLine{
		Method:        string(fields[0]),
		RequestTarget: string(fields[1]),
		HttpVersion:   string(versionParts[1]),
	}
	if !r.ValidMethod() {
		return nil, restOfMsg, ERROR_INVALID_HTTP_METHOD
	}

	return r, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	rl, _, err := parseRequestLine(data)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("parse request line failed"),
			err,
		)
	}
	return &Request{*rl}, nil
}
