package fesgo

import (
	"html/template"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Context) HTML(status int, html string) {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.W.WriteHeader(status)
	c.W.Write([]byte(html))
}

func (c *Context) HTMLTemplate(name string, data any, fileNames ...string) (err error) {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err = t.ParseFiles(fileNames...)
	if err != nil {
		return
	}
	err = t.Execute(c.W, data)
	if err != nil {
		c.W.Write([]byte(err.Error()))
		return
	}
	return
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) (err error) {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err = t.ParseGlob(pattern)
	if err != nil {
		return
	}
	err = t.Execute(c.W, data)
	if err != nil {
		c.W.Write([]byte(err.Error()))
		return
	}
	return
}
