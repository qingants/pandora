package mini

import (
	"log"
	"net/http"
)

type HandleFunc func(*Context)

type Engine struct {
	router *router
}

func NewEngine() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (e *Engine) AddRouter(method string, pattern string, handler HandleFunc) {
	log.Printf("Route %s - %s", method, pattern)
	e.router.addRoute(method, pattern, handler)
}

func (e *Engine) GET(pattern string, handler HandleFunc) {
	e.AddRouter("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandleFunc) {
	e.AddRouter("POST", pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)
	e.router.handle(c)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
