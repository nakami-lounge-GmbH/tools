package importer

import (
	"encoding/json"
	"fmt"
)

// Error represents an error object for the API
type Error struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ErrorList represents the ErrorList of all errors and waraning for the API
type ErrorList struct {
	Errors   []*Error `json:"errors"`
	Warnings []*Error `json:"warnings"`
}

func (l *ErrorList) String() string {
	b, err := l.MarshalJSON()

	if err != nil {
		return "Error getting string: " + err.Error()
	}
	return string(b)
}

// MarshalJSON marshales the ErrorList and adds some information
func (l *ErrorList) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Errors      []*Error `json:"errors"`
		Warnings    []*Error `json:"warnings"`
		HasAny      bool     `json:"has_any"`
		HasErrors   bool     `json:"has_errors"`
		HasWarnings bool     `json:"has_warnings"`
	}{
		Errors:      l.Errors,
		Warnings:    l.Warnings,
		HasAny:      l.HasAny(),
		HasErrors:   l.HasErrors(),
		HasWarnings: l.HasWarning(),
	})
}

// HasAny returns if the ErrorList has either errors or warnings
func (l *ErrorList) HasAny() bool {
	return l.HasErrors() || l.HasWarning()
}

// HasErrors returns if the ErrorList contains errors
func (l *ErrorList) HasErrors() bool {
	return len(l.Errors) > 0
}

// HasWarning returns if the ErrorList contains warnings
func (l *ErrorList) HasWarning() bool {
	return len(l.Warnings) > 0
}

// AddValidation adds a new validation error
func (l *ErrorList) AddValidation(line int, column int, header string, err error) {
	l.Errors = append(l.Errors, &Error{
		Message: fmt.Sprintf("error on line: '%d' col: '%d', header: '%s' error: '%v'\n", line, column, header, err),
	})
}

// AddErrorC adds a new error object to the ErrorList (with code)
func (l *ErrorList) AddErrorC(err error, code string) {
	l.Errors = append(l.Errors, &Error{Code: code, Message: err.Error()})
}

// AddErrorMsgC adds a new error with the message prefixed to the ErrorList (with code)
func (l *ErrorList) AddErrorMsgC(err error, code string, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Errors = append(l.Errors, &Error{Code: code, Message: fmt.Sprintf("%s :: %s", msg, err.Error())})
}

// AddError adds a new error object to the ErrorList
func (l *ErrorList) AddError(err error) {
	l.Errors = append(l.Errors, &Error{Message: err.Error()})
}

// AddErrorMsg adds a new error with the message prefixed to the ErrorList
func (l *ErrorList) AddErrorMsg(err error, msg string, args ...interface{}) *ErrorList {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Errors = append(l.Errors, &Error{Message: fmt.Sprintf("%s :: %s", msg, err.Error())})
	return l
}

// AddErrorString adds a new error string to the ErrorList
func (l *ErrorList) AddErrorString(msg string, args ...interface{}) *ErrorList {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Errors = append(l.Errors, &Error{Message: msg})
	return l
}

// AddErrorStringC adds a new error string to the ErrorList (with code)
func (l *ErrorList) AddErrorStringC(code string, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Errors = append(l.Errors, &Error{Message: msg, Code: code})
}

// WARNINGS

// AddWarnC adds a new warning object to the ErrorList (with code)
func (l *ErrorList) AddWarnC(err error, code string) {
	l.Warnings = append(l.Warnings, &Error{Code: code, Message: err.Error()})
}

// AddWarnMsgC adds a new warning with the message prefixed to the ErrorList (with code)
func (l *ErrorList) AddWarnMsgC(err error, code string, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Warnings = append(l.Warnings, &Error{Code: code, Message: fmt.Sprintf("%s :: %s", msg, err.Error())})
}

// AddWarn adds a new warning object to the ErrorList
func (l *ErrorList) AddWarn(err error) {
	l.Warnings = append(l.Warnings, &Error{Message: err.Error()})
}

// AddWarnMsg adds a new warning with the message prefixed to the ErrorList
func (l *ErrorList) AddWarnMsg(err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Warnings = append(l.Warnings, &Error{Message: fmt.Sprintf("%s :: %s", msg, err.Error())})
}

// AddWarnString adds a new warning string to the ErrorList
func (l *ErrorList) AddWarnString(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Warnings = append(l.Warnings, &Error{Message: msg})
}

// AddWarnStringC adds a new warning string to the ErrorList (with code)
func (l *ErrorList) AddWarnStringC(code string, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.Warnings = append(l.Warnings, &Error{Message: msg, Code: code})
}
