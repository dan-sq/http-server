package request

import (
	"bytes"
	"fmt"
	"io"
	"main/internal/headers"
	"strconv"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string
const (
	StateInit parserState = "init"
	StateDone parserState = "done"
	StateHeaders parserState = "headers"
	StateBody parserState = "body"
)

type Request struct {
	RequestLine RequestLine
	Headers *headers.Headers
	state parserState
	Body string
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

var SEPARATOR = []byte("\r\n")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorInvalidHttpVersion = fmt.Errorf("invalid http version")
var ErrorInvalidHttpMethod = fmt.Errorf("invalid http method")

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: headers.NewHeaders(),
		Body: "",
	}
}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, read, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if string(httpParts[0]) != "HTTP" || len(httpParts) != 2 {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine {
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}

	if rl.Method != strings.ToUpper(rl.Method) {
		return nil, read, ErrorInvalidHttpMethod
	}

	if rl.HttpVersion != "1.1" {
		return nil, read, ErrorInvalidHttpVersion
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}				

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders


		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				encoding, ok := r.Headers.Get("transfer-encoding")
				if encoding == "chunked" && ok {
					return 0, fmt.Errorf("chunked encoding not implemented")
				}
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			remaining := min(length - len(r.Body), len(currentData)) 
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("we programmed wrong")
		}
	}
	return read, nil
}


func (r *Request) done() bool {
	return r.state == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	idx := 0
	for !request.done() {
		if idx == len(buf) {
			biggerBuf := make([]byte, len(buf) * 2)
			copy(biggerBuf, buf)
			buf = biggerBuf
		}

		n, err := reader.Read(buf[idx:])
		if err != nil {
			return nil, err
		}

		idx += n
		parseN, err := request.parse(buf[:idx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[parseN:idx])
		idx -= parseN
	}

	return request, nil
}
