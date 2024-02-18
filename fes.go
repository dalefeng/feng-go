package fesgo

import (
	"fmt"
	"github.com/dalefeng/fesgo/render"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Engine struct {
	router
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
	pool       sync.Pool
}

func NewEngine() *Engine {
	engine := &Engine{
		router: router{},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)
}

func (e *Engine) SetHtmlTemplate(t *template.Template) {
	e.HTMLRender = render.HTMLRender{Template: t}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	e.httpRequestHandle(ctx, w, r)
	ctx.ClearContext()
	e.pool.Put(ctx)
}

func (e *Engine) httpRequestHandle(ctx *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.routerGroups {
		// 将分组截取
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node == nil || !node.isEnd {
			// 路由没匹配
			ctx.StatusCode = http.StatusNotFound
			group.MethodHandle(ctx, routerName, ANY, nil)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s %s not found - tree node", r.RequestURI, method)
			return
		}
		// 优先匹配 Any
		handleFunc, ok := group.handleFuncMap[node.routerName][ANY]
		if ok {
			group.MethodHandle(ctx, node.routerName, ANY, handleFunc)
			return
		}

		// method 匹配
		handleFunc, ok = group.handleFuncMap[node.routerName][method]
		if ok {
			group.MethodHandle(ctx, node.routerName, method, handleFunc)
			return
		}

		ctx.StatusCode = http.StatusMethodNotAllowed
		group.MethodHandle(ctx, node.routerName, ANY, nil)
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "%s %s not allowed", r.RequestURI, method)
		return
	}
	// 路由匹配失败
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s %s not found", r.RequestURI, method)
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
