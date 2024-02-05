package fesgo

import (
	"fmt"
	"github.com/dalefeng/fesgo/render"
	"html/template"
	"net/http"
	"net/url"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request

	engine     *Engine
	queryCache url.Values
	queryMap   map[string]map[string]string
}

func (c *Context) ClearContext() {
	c.queryCache = nil
}

func (c *Context) initQueryCache() {
	if c.queryCache != nil && c.queryMap != nil {
		return
	}
	if c.R == nil {
		c.queryCache = url.Values{}
		c.queryMap = make(map[string]map[string]string)
		return
	}
	c.queryCache = c.R.URL.Query()
	c.queryMap = ParseParamsMap(c.queryCache)
}

func (c *Context) GetDefaultQuery(key string, defaultValue string) string {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	if !ok {
		return defaultValue
	}
	return values[0]
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	values, ok := c.queryMap[key]
	return values, ok
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}
func (c *Context) GetQueryArr(key string) []string {
	c.initQueryCache()
	return c.queryCache[key]
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
func (c *Context) Template(name string, data any) {
	err := c.Render(http.StatusOK, &render.HTML{
		Name:       name,
		Data:       data,
		Template:   c.engine.HTMLRender.Template,
		IsTemplate: true,
	})
	if err != nil {
		c.Abort(err)
		return
	}
	return
}

func (c *Context) JSON(status int, data any) {
	err := c.Render(status, &render.Json{Data: data})
	if err != nil {
		c.Abort(err)
		return
	}
}

func (c *Context) XML(status int, data any) {
	err := c.Render(status, &render.Xml{Data: data})
	if err != nil {
		c.Abort(err)
		return
	}
}

func (c *Context) File(fileName string) {
	http.ServeFile(c.W, c.R, fileName)
}

func (c *Context) FileAttachment(filepath string, fileName string) {
	if IsASCII(fileName) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(fileName))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// FileFromFS filepath 相对文件系统的
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)
	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// Redirect 重定向
func (c *Context) Redirect(status int, url string) {
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("cannot redirect with status code %d", status))
	}
	http.Redirect(c.W, c.R, url, status)
}

func (c *Context) String(status int, format string, values ...any) {
	err := c.Render(status, &render.String{Format: format, Data: values})
	if err != nil {
		c.Abort(err)
		return
	}
}

func (c *Context) Render(statusCode int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(statusCode)
	return err
}

func (c *Context) Abort(err error) {
	c.W.WriteHeader(http.StatusInternalServerError)
	c.W.Write([]byte(err.Error()))
}
