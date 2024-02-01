package fesgo

import (
	"fmt"
	"github.com/dalefeng/fesgo/render"
	"html/template"
	"log"
	"net/http"
)

type Engine struct {
	router
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.httpRequestHandle(w, r)
}

func (e *Engine) httpRequestHandle(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{W: w, R: r}
	method := r.Method
	for _, group := range e.routerGroups {
		// 将分组截取
		routerName := SubStringLast(r.RequestURI, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node == nil || !node.isEnd {
			// 路由没匹配
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

		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "%s %s not allowed", r.RequestURI, method)
		return
	}
	// 路由匹配失败
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s %s not found", r.RequestURI, method)
}

func NewEngine(port string) *Engine {
	return &Engine{
		router: router{},
	}
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
