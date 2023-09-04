package cat

import (
	"fmt"
	"net/http"
	"path"
)

// Group 需要有访问 Router 的能力，为了方便，在 Group 中，保存一个指针，指向 Engine
// 整个框架的所有资源都是由 Engine 统一协调的，可以通过 Engine 间接地访问各种接口
type RouteGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouteGroup
	engine      *Engine
}

func (group *RouteGroup) Group(prefix string) *RouteGroup {
	engine := group.engine
	newGroup := &RouteGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}

	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouteGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouteGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := fmt.Sprintf("%s%s", group.prefix, comp)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouteGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodGet, pattern, handler)
}

func (group *RouteGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute(http.MethodPost, pattern, handler)
}

func (group *RouteGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(ctx *Context) {
		file := ctx.Param("filepath")

		if _, err := fs.Open(file); err != nil {
			ctx.SetStatusCode(http.StatusNoContent)
			return
		}

		fileServer.ServeHTTP(ctx.Writer, ctx.Req)
	}
}

func (group *RouteGroup) Static(relativePath, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filename")

	group.GET(urlPattern, handler)
}
