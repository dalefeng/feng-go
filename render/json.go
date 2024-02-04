package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data any
}

func (j *Json) Render(w http.ResponseWriter) error {
	j.WriterContentType(w)
	jsonData, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonData)
	return err
}

func (j *Json) WriterContentType(w http.ResponseWriter) {
	writerContentType(w, "application/json; charset=utf-8")
}
