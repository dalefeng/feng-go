package binding

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	if n == 0 {
		return ""
	}

	var b strings.Builder
	if err[0] != nil {
		fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
	}
	if n == 1 {
		return b.String()
	}
	for i := 1; i < n; i++ {
		if err[i] == nil {
			continue
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
	}
	return b.String()
}

type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	of := reflect.ValueOf(obj)
	elem := of.Elem()

	for elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	errs := make(SliceValidationError, 0)
	switch elem.Kind() {
	case reflect.Struct:
		return d.validateStruct(elem)
	case reflect.Slice, reflect.Array:
		count := elem.Len()
		for i := 0; i < count; i++ {
			err := d.validateStruct(elem.Index(i).Interface())
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs
		}
		return nil
	default:
		return nil
	}
}

func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}
func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}
