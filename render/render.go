package render

import "net/http"

type Render interface {
	Render(w http.ResponseWriter) error
	WriterContentType(w http.ResponseWriter)
}

func writerContentType(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", value)
}
