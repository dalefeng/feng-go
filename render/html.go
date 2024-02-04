package render

import (
	"github.com/dalefeng/fesgo/render/internal/bytescovn"
	"html/template"
	"net/http"
)

type HTML struct {
	Name       string
	Data       any
	Template   *template.Template
	IsTemplate bool
}

type HTMLRender struct {
	Template *template.Template
}

func (r *HTML) Render(w http.ResponseWriter) error {
	r.WriterContentType(w)
	if r.IsTemplate {
		err := r.Template.ExecuteTemplate(w, r.Name, r.Data)
		return err
	}
	_, err := w.Write(bytescovn.StringToBytes(r.Data.(string)))
	return err
}

func (r *HTML) WriterContentType(w http.ResponseWriter) {
	writerContentType(w, "text/html; charset=utf-8")
}
