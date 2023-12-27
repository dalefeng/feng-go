package fesgo

import (
	"fmt"
	"log"
	"net/http"
)

type Engine struct {
	router
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{W: w, R: r}
	method := r.Method
	for _, group := range e.routerGroups {
		// 将分组截取
		routerName := SubStringLast(r.RequestURI, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node == nil {
			// 路由没匹配
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s %s not found - tree node", r.RequestURI, method)
			return
		}
		// 优先匹配 Any
		handleFunc, ok := group.handleFuncMap[routerName][ANY]
		if ok {
			handleFunc(ctx)
			return
		}

		// method 匹配
		handleFunc, ok = group.handleFuncMap[routerName][method]
		if ok {
			handleFunc(ctx)
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
