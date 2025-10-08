package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"main/internal/headers"
	"main/internal/request"
	"main/internal/response"
	"main/internal/server"
)

const port = 42069

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)	
	}

	return out
}

func respond400() []byte{
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		if req.RequestLine.RequestTarget == "/yourprobelm" {
			status = response.StatusBadRequest
			body = respond400()

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = respond500()

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/video") {
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-type", "video/mp4")
			h.Replace("Content-length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(*h)
			w.WriteBody(f)

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget

			r, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				status = response.StatusInternalServerError
				body = respond500()
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("Content-length")
				h.Set("Transfer-encoding", "chunked")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				h.Replace("Content-type", "text/plain")
				w.WriteHeaders(*h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := r.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				trailer := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailer.Set("X-Content-SHA256", toStr(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailer)
				w.WriteBody([]byte("\r\n"))

				return
			}
		}
		
		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}