package render

import (
	"encoding/xml"
	"net/http"
)

type Xml struct {
	Data any
}

func (x *Xml) Render(w http.ResponseWriter) error {
	x.WriterContentType(w)
	err := xml.NewEncoder(w).Encode(x.Data)
	return err
}

func (x *Xml) WriterContentType(w http.ResponseWriter) {
	writerContentType(w, "application/xml; charset=utf-8")
}
