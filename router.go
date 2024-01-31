package fesgo

import (
	"net/http"
)

const ANY = "ANY"

type HandlerFunc func(ctx *Context)

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		name:            name,
		handleFuncMap:   make(map[string]map[string]HandlerFunc),
		handleMethodMap: make(map[string][]string),
		treeNode:        &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, group)

	return group
}

type routerGroup struct {
	name            string // 组名
	handleFuncMap   map[string]map[string]HandlerFunc
	handleMethodMap map[string][]string
	treeNode        *treeNode
	preMiddleware   []MiddlewareFunc
	postMiddleware  []MiddlewareFunc
}

func (r *routerGroup) PreHandle(middlewareFunc ...MiddlewareFunc) {
	r.preMiddleware = append(r.preMiddleware, middlewareFunc...)
}

func (r *routerGroup) PostHandle(middlewareFunc ...MiddlewareFunc) {
	r.postMiddleware = append(r.preMiddleware, middlewareFunc...)
}

func (r *routerGroup) MethodHandle(ctx *Context, h HandlerFunc) {
	// 前置中间件
	if r.preMiddleware != nil {
		for _, middlewareFunc := range r.preMiddleware {
			h = middlewareFunc(h)
		}
	}
	h(ctx)
	// 后置中间件
	if r.postMiddleware != nil {
		for _, middlewareFunc := range r.postMiddleware {
			h = middlewareFunc(h)
		}
	}
	h(ctx)
}

func (r *routerGroup) handle(name, method string, handleFunc HandlerFunc) {
	_, ok := r.handleFuncMap[name]
	if !ok {
		r.handleFuncMap[name] = make(map[string]HandlerFunc)
	}
	_, ok = r.handleFuncMap[name][method]
	if ok {
		panic("路由已存在")
	}
	r.handleFuncMap[name][method] = handleFunc

	r.treeNode.Put(name)
}

func (r *routerGroup) Any(name string, handlerFunc HandlerFunc) {
	r.handle(name, ANY, handlerFunc)
}
func (r *routerGroup) Get(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodGet, handlerFunc)
}
func (r *routerGroup) Post(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPost, handlerFunc)
}
func (r *routerGroup) Put(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPut, handlerFunc)
}
func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodDelete, handlerFunc)
}
func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPatch, handlerFunc)
}
func (r *routerGroup) Head(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodHead, handlerFunc)
}
