package router

import (
	"main/internal/request"
	"main/internal/response"
	"main/internal/server"
)

type Router struct {
	routes map[string]map[string]server.Handler
}

func NewRouter() *Router {
	return &Router{ routes: map[string]map[string]server.Handler{} }
}

func (r *Router) Add(method string, path string, handler server.Handler) {
	if r.routes[path] == nil {
		r.routes[path] = map[string]server.Handler{}
	}
	r.routes[path][method] = handler
}

func (r *Router) Handle(w *response.Writer, req *request.Request) {
	methods, ok:= r.routes[req.RequestLine.RequestTarget]
	if !ok {
		w.WriteStatusLine(response.StatusNotFound)
		w.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}
	
	handler, ok := methods[req.RequestLine.Method]
	if !ok {
		w.WriteStatusLine(response.StatusMethodNotAllowed)
		w.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	handler(w, req)
}
