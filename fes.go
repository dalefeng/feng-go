package fesgo

import (
	"log"
	"net/http"
)

type Engine struct {
	router
}

func NewEngine(port string) *Engine {
	return &Engine{
		router: router{},
	}
}

func (e *Engine) Run() {
	for _, group := range e.routerGroups {
		for path, handlefunc := range group.handleFuncMap {
			http.HandleFunc("/"+group.name+path, handlefunc)
		}
	}

	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
