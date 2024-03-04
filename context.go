package fesgo

import (
	"errors"
	"fmt"
	"github.com/dalefeng/fesgo/binding"
	fesLog "github.com/dalefeng/fesgo/logger"
	"github.com/dalefeng/fesgo/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sync"
)

var defaultMultipartMemory int64 = 32 << 20 // 32M

type Context struct {
	W http.ResponseWriter
	R *http.Request

	engine     *Engine
	queryCache url.Values
	queryMap   map[string]map[string]string
	formCache  url.Values
	formMap    map[string]map[string]string

	StatusCode            int
	DisallowUnknownFields bool
	Logger                *fesLog.Logger

	Keys map[string]any
	mu   sync.RWMutex

	sameSite http.SameSite
}

func (c *Context) SetSameSite(s http.SameSite) {
	c.sameSite = s
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value string, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.Keys[key].(string)
	return
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

func (c *Context) initFormCache() {
	if c.formCache != nil && c.R.PostForm != nil {
		return
	}
	if c.R == nil {
		c.formCache = url.Values{}
		c.formMap = make(map[string]map[string]string)
		return
	}
	if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			log.Println(err)
		}
	}
	c.formCache = c.R.PostForm
	c.formMap = ParseParamsMap(c.formCache)
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

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	values, ok := c.formMap[key]
	return values, ok
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}
func (c *Context) GetPostForm(key string) string {
	c.initFormCache()
	return c.formCache.Get(key)
}
func (c *Context) GetQueryArray(key string) []string {
	c.initQueryCache()
	return c.queryCache[key]
}

func (c *Context) GetPostFormArray(key string) []string {
	c.initFormCache()
	return c.queryCache[key]
}

func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

func (c *Context) FormFiles(name string) []*multipart.FileHeader {
	forms, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0)
	}
	return forms.File[name]
}

func (c *Context) SaveUploadFile(file *multipart.FileHeader, dstPath string) (err error) {
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

func (c *Context) BindJson(obj any) error {
	json := binding.JSON
	json.DisallowUnknownFields = false
	return c.MustBindWith(obj, json)
}

func (c *Context) MustBindWith(obj any, bind binding.Binding) error {
	err := c.ShouldBind(obj, bind)
	if err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBind(obj any, binding binding.Binding) error {
	return binding.Bind(c.R, obj)
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
	if statusCode == http.StatusOK {
		c.StatusCode = http.StatusOK
	} else {
		c.SetStatusCode(statusCode)
	}
	err := r.Render(c.W)

	return err
}

func (c *Context) Abort(err error) {
	c.SetStatusCode(http.StatusInternalServerError)
	c.W.Write([]byte(err.Error()))
}

func (c *Context) SetStatusCode(code int) {
	c.W.WriteHeader(code)
	c.StatusCode = code
}

func (c *Context) HandleError(statusCode int, obj any, err error) {
	if err != nil {
		code, data := c.engine.errorHandler(err)
		c.JSON(code, data)
		return
	}
	c.JSON(statusCode, obj)
}

func (c *Context) SetBase64Auth(username, password string) {
	c.R.Header.Set("Authorization", "Basic "+BasicAuth(username, password))
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	http.SetCookie(c.W, cookie)
}
