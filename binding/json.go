package binding

import (
	"encoding/json"
	"errors"
	"net/http"
)

var Validate StructValidator = &defaultValidator{}

type JsonBinding struct {
	DisallowUnknownFields bool
}

func (j JsonBinding) Name() string {
	return "json"
}

func (j JsonBinding) Bind(r *http.Request, obj any) error {
	body := r.Body
	defer body.Close()
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if j.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	return Validate.ValidateStruct(obj)
}
