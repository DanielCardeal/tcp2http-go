package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var LINE_SEPARATOR = []byte("\r\n")

var ErrorMalformedFieldLine = fmt.Errorf("malformed field-line")
var InvalidFieldName = fmt.Errorf("invalid field name")

type Headers struct {
	headers map[string]string
}

type HeaderEntry struct {
	Name  string
	Value string
}

func NewHeaders() *Headers {
	return &Headers{headers: make(map[string]string)}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Entries() []HeaderEntry {
	l := make([]HeaderEntry, 0, len(h.headers))
	for k, v := range h.headers {
		l = append(l, HeaderEntry{Name: k, Value: v})
	}
	return l
}

func (h *Headers) set(name, value string) {
	name = strings.TrimSuffix(name, ":")
	name = strings.ToLower(name)

	if prev := h.Get(name); len(prev) != 0 {
		h.headers[name] = fmt.Sprintf("%s,%s", prev, value)
	} else {
		h.headers[name] = value
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	idx := bytes.Index(data, LINE_SEPARATOR)
	switch idx {
	case -1:
		return 0, false, nil
	case 0:
		return len(LINE_SEPARATOR), true, nil
	}

	fields := bytes.Fields(data[:idx])
	if len(fields) != 2 {
		return 0, false, ErrorMalformedFieldLine
	}
	if !validFieldName(string(fields[0])) {
		return 0, false, InvalidFieldName
	}

	h.set(string(fields[0]), string(fields[1]))
	return idx + len(LINE_SEPARATOR), false, nil
}

func validFieldName(b string) bool {
	if len(b) <= 1 || !strings.HasSuffix(b, ":") {
		return false
	}

	for _, ch := range b[:len(b)-1] {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
			continue
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		}
		return false
	}
	return true
}
