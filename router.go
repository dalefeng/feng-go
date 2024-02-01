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
		name:              name,
		handleFuncMap:     make(map[string]map[string]HandlerFunc),
		middlewareFuncMap: make(map[string]map[string][]MiddlewareFunc),
		handleMethodMap:   make(map[string][]string),
		treeNode:          &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, group)

	return group
}

type routerGroup struct {
	name              string                                 // 组名
	handleFuncMap     map[string]map[string]HandlerFunc      // map[routerName]map[methods]HandlerFunc
	middlewareFuncMap map[string]map[string][]MiddlewareFunc //  map[routerName]map[methods][]路由级中间件

	handleMethodMap map[string][]string
	treeNode        *treeNode
	middleware      []MiddlewareFunc // 组级中间件
}

func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middlewareFunc...)
}

func (r *routerGroup) MethodHandle(ctx *Context, name, method string, h HandlerFunc) {
	// 组中间件
	if r.middleware != nil {
		for _, middlewareFunc := range r.middleware {
			h = middlewareFunc(h)
		}
	}
	// 路由级别
	middlewareFunc, ok := r.middlewareFuncMap[name][method]
	if ok {
		for _, mFunc := range middlewareFunc {
			h = mFunc(h)
		}
	}
	h(ctx)
}

func (r *routerGroup) handle(name, method string, handleFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	_, ok := r.handleFuncMap[name]
	if !ok {
		r.handleFuncMap[name] = make(map[string]HandlerFunc)
		r.middlewareFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	_, ok = r.handleFuncMap[name][method]
	if ok {
		panic("路由已存在")
	}
	r.handleFuncMap[name][method] = handleFunc
	r.middlewareFuncMap[name][method] = append(r.middlewareFuncMap[name][method], middlewareFunc...)

	r.treeNode.Put(name)
}

func (r *routerGroup) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, ANY, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Get(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodGet, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Post(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPost, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Put(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPatch, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Head(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodHead, handlerFunc, middlewareFunc...)
}
