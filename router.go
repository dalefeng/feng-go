package fesgo

import (
	"net/http"
)

const ANY = "ANY"

type HandleFunc func(ctx *Context)

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		name:            name,
		handleFuncMap:   make(map[string]map[string]HandleFunc),
		handleMethodMap: make(map[string][]string),
	}
	r.routerGroups = append(r.routerGroups, group)

	return group
}

type routerGroup struct {
	name            string // 组名
	handleFuncMap   map[string]map[string]HandleFunc
	handleMethodMap map[string][]string
}

func (r *routerGroup) handle(name, method string, handleFunc HandleFunc) {
	_, ok := r.handleFuncMap[name]
	if !ok {
		r.handleFuncMap[name] = make(map[string]HandleFunc)
	}
	_, ok = r.handleFuncMap[name][method]
	if ok {
		panic("路由已存在")
	}
	r.handleFuncMap[name][method] = handleFunc
	r.handleMethodMap[method] = append(r.handleMethodMap[method], name)
}

func (r *routerGroup) Any(name string, handleFunc HandleFunc) {
	r.handle(name, ANY, handleFunc)
}
func (r *routerGroup) Get(name string, handleFunc HandleFunc) {
	r.handle(name, http.MethodGet, handleFunc)
}
func (r *routerGroup) Post(name string, handleFunc HandleFunc) {
	r.handle(name, http.MethodPost, handleFunc)
}
func (r *routerGroup) Put(name string, handleFunc HandleFunc) {
	r.handle(name, http.MethodPut, handleFunc)
}
func (r *routerGroup) Delete(name string, handleFunc HandleFunc) {
	r.handle(name, http.MethodDelete, handleFunc)
}
