package response

import (
	"fmt"
	"io"

	"main/internal/headers"
)

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	b := []byte {}
	headers.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)

	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case StatusOK: statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest: statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError: statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default: 
		return fmt.Errorf("unrecognized error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}


func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}