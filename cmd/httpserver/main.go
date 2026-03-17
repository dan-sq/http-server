package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"main/internal/request"
	"main/internal/response"
	"main/internal/router"
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

func sayHello(w *response.Writer, req *request.Request) {
	body := respond200()
	h := response.GetDefaultHeaders(len(body))
	status := response.StatusOK
	
	w.WriteStatusLine(status)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func main() {
	r := router.NewRouter()
	r.Add("GET", "/api/hello", sayHello)

	s, err := server.Serve(port, r.Handle)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<- sigChan
	log.Println("Server gracefully stopped")
}
