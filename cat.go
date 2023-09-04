package cat

import (
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by cat
type HandlerFunc func(ctx *Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouteGroup
	router *router
	groups []*RouteGroup // store all groups
}

// New is the constructor of cat.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouteGroup = &RouteGroup{
		engine: engine,
	}
	engine.groups = []*RouteGroup{engine.RouteGroup}
	return engine
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares = make([]HandlerFunc, 0)

	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)
}
