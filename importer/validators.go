package importer

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

// validates, that at least one field is set
func oneFieldSet(fl validator.FieldLevel) bool {
	field := fl.Field()

	currentField, _, ok := fl.GetStructFieldOK()
	if !ok {
		return false
	}

	defaultF := field.Interface() == reflect.Zero(field.Type()).Interface()
	defaultC := currentField.Interface() == reflect.Zero(currentField.Type()).Interface()

	if !defaultF || !defaultC {
		return true
	}

	return false
}

// validates, that at if field is set, the other one is also set
func withFieldSet(fl validator.FieldLevel) bool {
	field := fl.Field()

	currentField, _, ok := fl.GetStructFieldOK()
	if !ok {
		return false
	}

	defaultF := field.Interface() == reflect.Zero(field.Type()).Interface()
	defaultC := currentField.Interface() == reflect.Zero(currentField.Type()).Interface()

	if defaultF || (!defaultF && !defaultC) {
		return true
	}

	return false
}

// oneOfStr, checks that the value is one of string with special separator
func oneOfStr(fl validator.FieldLevel) bool {
	fieldVal := strings.ToLower(fl.Field().String())

	if fieldVal == "" {
		return true
	}

	param := fl.Param()
	if param == "" {
		return true
	}

	params := strings.Split(param, "#-#")
	for _, s := range params {
		if strings.ToLower(s) == fieldVal {
			return true
		}
	}

	return false
}
