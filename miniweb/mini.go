package miniweb

import (
	"log"
	"net/http"
)

type HandleFunc func(*Context)

type RouterGroup struct {
	prefix string
	parent *RouterGroup
	engine *Engine
}

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func NewEngine() *Engine {
	e := &Engine{
		router: newRouter(),
	}
	e.RouterGroup = &RouterGroup{engine: e}
	e.groups = []*RouterGroup{e.RouterGroup}
	return e
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix: g.prefix + "/" + prefix,
		engine: engine,
		parent: g,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (g *RouterGroup) addRouter(method string, pattern string, handler HandleFunc) {
	pattern = g.prefix + pattern
	log.Printf("Route %4s - %s", method, pattern)
	g.engine.router.addRoute(method, pattern, handler)
}

func (g *RouterGroup) GET(pattern string, handler HandleFunc) {
	g.addRouter("GET", pattern, handler)
}

func (g *RouterGroup) POST(pattern string, handler HandleFunc) {
	g.addRouter("POST", pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)
	e.router.handle(c)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
