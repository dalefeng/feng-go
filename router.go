package fesgo

import "net/http"

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type routerGroup struct {
	name          string // 组名
	handleFuncMap map[string]HandleFunc
}

func (r *routerGroup) Add(name string, handleFunc HandleFunc) {
	r.handleFuncMap[name] = handleFunc
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		name:          name,
		handleFuncMap: make(map[string]HandleFunc),
	}
	r.routerGroups = append(r.routerGroups, group)

	return group
}
