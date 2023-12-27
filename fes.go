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
		for name, handleFuncMap := range group.handleFuncMap {
			url := "/" + group.name + name
			// 路由匹配
			if r.RequestURI != url {
				continue
			}

			// 优先匹配 Any
			handleFunc, ok := handleFuncMap[ANY]
			if ok {
				handleFunc(ctx)
				return
			}

			// method 匹配
			handleFunc, ok = handleFuncMap[method]
			if ok {
				handleFunc(ctx)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed", r.RequestURI, method)
			return
		}
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
