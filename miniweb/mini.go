package miniweb

import (
	"log"
	"net/http"
	"path"
	"strings"
)

type HandleFunc func(*Context)

type RouterGroup struct {
	prefix      string
	parent      *RouterGroup
	engine      *Engine
	middlewares []HandleFunc
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

func (g *RouterGroup) Use(middlewares ...HandleFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
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

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absPath := path.Join(g.prefix, relativePath)
	fileSever := http.StripPrefix(absPath, http.FileServer(fs))
	return func(ctx *Context) {
		file := ctx.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		fileSever.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (g *RouterGroup) Static(relativePath, root string) {
	hander := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	g.GET(urlPattern, hander)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandleFunc
	for _, group := range e.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := NewContext(w, r)
	c.handlers = middlewares
	e.router.handle(c)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
