package importer

import (
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestNullableFromLine covers nullable fields on the CSV path. The types of
// database/sql have the same shape as the null-types of our older projects, so
// they stand in for them here without pulling in a dependency
func TestNullableFromLine(t *testing.T) {
	type row struct {
		Name   sql.NullString  `header:"Name"`
		Count  sql.NullInt64   `header:"Anzahl"`
		Amount sql.NullFloat64 `header:"Betrag"`
		Active sql.NullBool    `header:"Aktiv"`
		Day    sql.NullTime    `header:"Datum" format:"02.01.2006"`
		Empty  sql.NullString  `header:"Leer"`
		Plain  int             `header:"Plain"`
	}

	headers := []string{"Name", "Anzahl", "Betrag", "Aktiv", "Datum", "Leer", "Plain"}

	tests := []struct {
		name string
		data []string
		want row
	}{
		{
			name: "values set",
			data: []string{"Polder", "1.234", "777,7", "ja", "02.10.1975", "", "42"},
			want: row{
				Name:   sql.NullString{String: "Polder", Valid: true},
				Count:  sql.NullInt64{Int64: 1234, Valid: true},
				Amount: sql.NullFloat64{Float64: 777.7, Valid: true},
				Active: sql.NullBool{Bool: true, Valid: true},
				Day:    sql.NullTime{Time: time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC), Valid: true},
				Plain:  42,
			},
		},
		{
			//a false bool is still a value: Valid must be set
			name: "false is a value, empty is not",
			data: []string{"", "", "", "nein", "", "", "0"},
			want: row{Active: sql.NullBool{Bool: false, Valid: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := new(ErrorList)
			ii := newImorter[row](&Config{dataType: ExcelLine}, el)
			ii.AddHeaders(headers)
			if el.HasAny() {
				t.Fatalf("header error: %s", el.String())
			}

			mV := ii.GetLineValues(1, tt.data)
			if el.HasAny() {
				t.Fatalf("import error: %s", el.String())
			}
			if mV == nil {
				t.Fatal("no values returned")
			}

			if got := mV.Interface().(row); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mismatch\n got: %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

// TestNullableFromExcel covers nullable fields on the Excel path
func TestNullableFromExcel(t *testing.T) {
	type data struct {
		Name   sql.NullString  `excel_pos:"B4"`
		Day    sql.NullTime    `excel_pos:"E4"`
		Kids   sql.NullInt64   `excel_pos:"F4"`
		Amount sql.NullFloat64 `excel_pos:"H5"`
		//H4 does not exist in the file
		Empty sql.NullFloat64 `excel_pos:"H4"`
		//C4 holds text, so a date must stay unset instead of being set to zero
		NotADate sql.NullTime `excel_pos:"C4"`
	}

	el := new(ErrorList)
	i, err := NewExcelPageImporter[data](&ExcelPageConfig{
		Config{FileBytes: readFixture(t), SheetName: "Daten"},
	}, el)

	if err != nil {
		t.Fatalf("with error: %v errorlist: %s", err, el.String())
	}
	if el.HasAny() {
		t.Fatal(el.String())
	}

	want := []data{{
		Name:   sql.NullString{String: "Polder", Valid: true},
		Day:    sql.NullTime{Time: time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC), Valid: true},
		Kids:   sql.NullInt64{Int64: 4, Valid: true},
		Amount: sql.NullFloat64{Float64: 777.7, Valid: true},
	}}
	if !reflect.DeepEqual(i.Data, want) {
		t.Errorf("mismatch\n got: %+v\nwant: %+v", i.Data, want)
	}
}

// TestNullableGuard makes sure the shape check does not accept any struct it
// comes across: a field it cannot fill must still be reported
func TestNullableGuard(t *testing.T) {
	type threeFields struct {
		Value string
		Valid bool
		More  int
	}
	type validNotBool struct {
		Value string
		Valid int
	}
	type flagNotNamedValid struct {
		Value string
		Set   bool
	}
	type valueFirstMissing struct {
		Valid bool
		Value string
	}

	tests := []struct {
		name string
		run  func(eL *ErrorList) error
	}{
		{"three fields", func(eL *ErrorList) error {
			type data struct {
				F threeFields `excel_pos:"B4"`
			}
			_, err := newPageImporter[data](eL)
			return err
		}},
		{"Valid is not a bool", func(eL *ErrorList) error {
			type data struct {
				F validNotBool `excel_pos:"B4"`
			}
			_, err := newPageImporter[data](eL)
			return err
		}},
		{"flag not named Valid", func(eL *ErrorList) error {
			type data struct {
				F flagNotNamedValid `excel_pos:"B4"`
			}
			_, err := newPageImporter[data](eL)
			return err
		}},
		{"value is not the first field", func(eL *ErrorList) error {
			type data struct {
				F valueFirstMissing `excel_pos:"B4"`
			}
			_, err := newPageImporter[data](eL)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := new(ErrorList)
			if err := tt.run(el); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !el.HasErrors() {
				t.Fatal("want an error for a field that cannot be filled, got none")
			}
			if !strings.Contains(el.String(), "undefined kind") {
				t.Errorf("want an 'undefined kind' error, got: %s", el.String())
			}
		})
	}
}

// TestNullableAcceptsStdlibShape documents which shapes count as nullable
func TestNullableAcceptsStdlibShape(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"sql.NullString", sql.NullString{}, true},
		{"sql.NullTime", sql.NullTime{}, true},
		{"sql.NullInt64", sql.NullInt64{}, true},
		{"sql.NullFloat64", sql.NullFloat64{}, true},
		{"sql.NullBool", sql.NullBool{}, true},
		{"time.Time", time.Time{}, false},
		{"plain struct", struct {
			A string
			B bool
		}{}, false},
		{"unexported fields", struct {
			value string
			Valid bool
		}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//the field must be addressable to be settable, as it is in an import struct
			v := reflect.New(reflect.TypeOf(tt.value)).Elem()
			if _, ok := nullableValue(v); ok != tt.want {
				t.Errorf("nullableValue = %v, want %v", ok, tt.want)
			}
		})
	}
}

// newPageImporter is a short way to run a page import against the fixture
func newPageImporter[C any](eL *ErrorList) (ExcelPageImporter[C], error) {
	fileBytes, err := os.ReadFile("test1.xlsx")
	if err != nil {
		return ExcelPageImporter[C]{}, err
	}

	return NewExcelPageImporter[C](&ExcelPageConfig{
		Config{FileBytes: fileBytes, SheetName: "Daten"},
	}, eL)
}
