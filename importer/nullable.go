package importer

import (
	"reflect"
)

// nullValidFieldName is the name of the flag field a nullable type carries
const nullValidFieldName = "Valid"

// nullableValue reports whether f is a nullable field and returns the field
// holding its value.
//
// A nullable is a struct of the shape "{Value T; Valid bool}": a value together
// with a flag telling whether it is set. The null-types used in some of our
// older projects are built that way, as are the Null-types of the standard
// library's database/sql. The importer fills them through their value field, so
// it does not need to know the package they come from
func nullableValue(f reflect.Value) (reflect.Value, bool) {
	if f.Kind() != reflect.Struct || f.NumField() != 2 {
		return reflect.Value{}, false
	}

	valid := f.Type().Field(1)
	if valid.Name != nullValidFieldName || valid.Type.Kind() != reflect.Bool {
		return reflect.Value{}, false
	}

	if !f.Field(0).CanSet() || !f.Field(1).CanSet() {
		return reflect.Value{}, false
	}

	return f.Field(0), true
}

// setNullableValid marks a nullable field as set. It must only be called for a
// field nullableValue has accepted
func setNullableValid(f reflect.Value) {
	f.Field(1).SetBool(true)
}
