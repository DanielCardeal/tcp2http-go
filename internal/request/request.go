package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/DanielCardeal/tcp2http-go/internal/headers"
)

type ParserState int

const (
	StateInitialized ParserState = iota
	StateHeaders
	StateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
}

type requestParser struct {
	request *Request
	state   ParserState

	data []byte
	pos  int
}

var SEPARATOR = []byte("\r\n")

var ErrorIncompleteRequest = fmt.Errorf("incomplete request")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHTTPVersion = fmt.Errorf("unsupported HTTP version")
var ErrorInvalidHTTPMethod = fmt.Errorf("unknown HTTP method")
var ErrorUnknownParserState = fmt.Errorf("unknown parser state")

func validHttpMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "DELETE":
		return true
	default:
		return false
	}
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
		return nil, 0, ErrorMalformedRequestLine
	}

	versionParts := bytes.Split(fields[2], []byte("/"))
	if len(versionParts) != 2 ||
		string(versionParts[0]) != "HTTP" ||
		string(versionParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	if !validHttpMethod(string(fields[0])) {
		return nil, 0, ErrorInvalidHTTPMethod
	}

	r := &RequestLine{
		Method:        string(fields[0]),
		RequestTarget: string(fields[1]),
		HttpVersion:   string(versionParts[1]),
	}

	return r, read, nil
}

func newRequest() *Request {
	return &Request{Headers: *headers.NewHeaders()}
}

func newRequestParser() *requestParser {
	return &requestParser{
		request: newRequest(),
		state:   StateInitialized,
		data:    make([]byte, 0, 4096),
		pos:     0,
	}
}

func (p *requestParser) finished() bool {
	return p.state == StateDone
}

func (p *requestParser) parse(newData []byte) error {
	p.data = append(p.data, newData...)

	for {
		switch p.state {
		case StateInitialized:
			rl, n, err := parseRequestLine(p.data[p.pos:])
			if err != nil || n == 0 {
				return err
			}
			p.request.RequestLine = *rl
			p.state = StateHeaders
			p.pos += n
		case StateHeaders:
			n, done, err := p.request.Headers.Parse(p.data[p.pos:])
			if err != nil || n == 0 {
				return err
			}
			p.pos += n
			if done {
				p.state = StateDone
			}
		case StateDone:
			return nil
		default:
			return ErrorUnknownParserState
		}
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readBuff := make([]byte, 4096)
	reqParser := newRequestParser()
	for !reqParser.finished() {
		n, err := reader.Read(readBuff)
		if err != nil {
			return nil, ErrorIncompleteRequest
		}
		if n == 0 {
			continue
		}

		err = reqParser.parse(readBuff[:n])
		if err != nil {
			return nil, err
		}
	}
	return reqParser.request, nil
}
