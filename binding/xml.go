package binding

import (
	"encoding/xml"
	"errors"
	"net/http"
)

type XmlBinding struct {
}

func (x XmlBinding) Name() string {
	return "xml"
}

func (x XmlBinding) Bind(r *http.Request, obj any) error {
	body := r.Body
	defer body.Close()
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := xml.NewDecoder(body)

	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	return Validate.ValidateStruct(obj)
}
