package importer

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/nakami-lounge-GmbH/tools/importer/excel_cols"
	"golang.org/x/exp/slices"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ImportType int

const (
	ExcelLine    ImportType = iota
	ExcelPage    ImportType = iota
	ExcelLineCol ImportType = iota
)

const (
	tagNameLine       = "header"
	tagNameLineCol    = "col"
	tagNamePage       = "excel_pos"
	tagNameDateFormat = "format"
)

var (
	//ErrNotAllCols is returned, if not all rows on the import have the requiered columns
	ErrNotAllCols = errors.New("not all columns set")

	//validate is used for validating the structures
	validate *validator.Validate
)

type Config struct {
	dataType          ImportType
	SheetName         string   //either name or number (0-indexed)
	SheetNumber       int      //either name or number (0-indexed)
	OffsetRow         int      //this should be the header row (not 0-indexed)
	FileBytes         []byte   //data from the excel file
	LineCountToRead   int      //specifies, how many lines to read
	EmptyValueStrings []string //specifies values that should be treated as empty
}

type Importer[C any] struct {
	Config    *Config
	errorList *ErrorList

	//Fields holds the position inside/col the data and also an reflect-StructField of the field itself
	//We use it to create the data, when filling it to the final struct
	fields map[string]*field

	//DataHeaders are the headers that are defined in the data to be imported
	dataHeaders []string

	//MinColCount is the minumum number of cols, every data-object must have.
	//We use this, for checking, that the input data has at least the requiered amount of cols
	minColCount int

	//StructType is the type of the final struct
	structType reflect.Type

	//will hold the final data
	Data []C
}

// GetSheet reads the sheet named in the config, or, if no name is set, the
// sheet with the configured 1-based number. The caller must Close the sheet
func (ii *Importer[C]) GetSheet() (*Sheet, error) {
	if ii.Config.SheetName == "" && ii.Config.SheetNumber == 0 {
		return nil, errors.New("no sheet selected: set either SheetName or SheetNumber")
	}

	return openSheet(ii.Config.FileBytes, ii.Config.SheetName, ii.Config.SheetNumber)
}

type field struct {
	PosInData int    //used in line import. Specifies the position in the data
	Col       string //used in line import. Specifies the position in the data
	ExcelPos  string //used in page import. Holds the Excel position from the Tag
	Field     reflect.StructField
}

func newImorter[C any](config *Config, eL *ErrorList) Importer[C] {
	var ii Importer[C]
	ii.Config = config
	ii.errorList = eL

	ii.fields = make(map[string]*field)
	validate = validator.New()
	_ = validate.RegisterValidation("eitherrequired", oneFieldSet)
	_ = validate.RegisterValidation("withrequired", withFieldSet)
	_ = validate.RegisterValidation("one_of_str", oneOfStr)

	cc := new(C)
	tt := reflect.TypeOf(cc)
	if tt.Kind() != reflect.Ptr {
		eL.AddErrorString("Only pointer type allowed")
		return ii
	}

	v := reflect.ValueOf(cc)
	if v.IsNil() {
		ii.errorList.AddErrorString("No nill pointer allowed")
		return ii
	}

	t := tt.Elem()
	ii.structType = t

	tagName := tagNameLine
	if ii.Config.dataType == ExcelPage {
		tagName = tagNamePage
	} else if ii.Config.dataType == ExcelLineCol {
		tagName = tagNameLineCol
	}

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get(tagName)
		if tag != "" {
			if ii.Config.dataType == ExcelLine {
				ii.fields[strings.ToLower(tag)] = &field{Field: f, PosInData: -1}
			} else if ii.Config.dataType == ExcelPage {
				ii.fields[strings.ToLower(tag)] = &field{Field: f, ExcelPos: tag}
			} else if ii.Config.dataType == ExcelLineCol {
				ii.fields[strings.ToLower(tag)] = &field{Field: f, Col: tag}
			}
		}
	}

	return ii
}

func (ii *Importer[C]) AddHeaders(AllHeaders []string) {
	//look through all headers and save the position
	for i, h := range AllHeaders {
		if f, ok := ii.fields[strings.ToLower(h)]; ok {
			f.PosInData = i
		}
	}

	ii.dataHeaders = AllHeaders

	ii.minColCount = -1
	for _, value := range ii.fields {
		if value.PosInData > ii.minColCount {
			ii.minColCount = value.PosInData
		}
	}

	//check header count
	for key, value := range ii.fields {
		if value.PosInData == -1 {
			ii.errorList.AddErrorString("Field '%s' is missing on import", key)
		}
	}
}

// GetLineValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
// GetLineValues shold be used for CSV imports, not Excel
func (ii *Importer[C]) GetLineValues(line int, data []string) *reflect.Value {
	v := reflect.New(ii.structType).Elem()

	if len(data) != len(ii.dataHeaders) {
		ii.errorList.AddValidation(line, 0, "", fmt.Errorf("Number of columns does not match. want %d have %d", len(ii.dataHeaders), len(data)))
		return nil
	}

	if len(data)-1 < ii.minColCount {
		ii.errorList.AddValidation(line, 0, "", ErrNotAllCols)
		return nil
	}

	for header, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		col := value.PosInData

		if data[col] != "" && !slices.Contains(ii.Config.EmptyValueStrings, data[col]) {
			if err := ii.setLineValue(f, data[col], value, header); err != nil {
				ii.errorList.AddValidation(line, col, header, err)
				return nil
			}
		}
	}

	return &v
}

// setLineValue converts one value of a data line into the struct field it
// belongs to. Nullable fields are filled through their value field and are
// marked as set afterwards, see nullableValue
func (ii *Importer[C]) setLineValue(f reflect.Value, s string, value *field, header string) error {
	switch f.Interface().(type) {
	case int, int32, int64:
		i, err := strconv.Atoi(cleanNumberString(s))
		if err != nil {
			return err
		}
		f.SetInt(int64(i))

	case uint, uint32, uint64:
		i, err := strconv.ParseUint(cleanNumberString(s), 10, f.Type().Bits())
		if err != nil {
			return err
		}
		f.SetUint(i)

	case string:
		f.SetString(s)

	case time.Time:
		//check if format-tag set
		format := value.Field.Tag.Get(tagNameDateFormat)
		if format == "" {
			return fmt.Errorf("No format tag on structfield '%s'", header)
		}

		t, err := time.Parse(format, s)
		if err != nil {
			return fmt.Errorf("Could not parse value '%s' as date-time with format '%s'", s, format)
		}
		f.Set(reflect.ValueOf(t))

	case float32, float64:
		i, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64) //we need to have only one format
		if err != nil {
			return err
		}
		f.SetFloat(i)

	case bool:
		f.SetBool(isTrueString(s))

	default:
		inner, ok := nullableValue(f)
		if !ok {
			return fmt.Errorf("undefined kind '%v' on structfield '%s'", f.Kind(), header)
		}

		if err := ii.setLineValue(inner, s, value, header); err != nil {
			return err
		}
		setNullableValid(f)
	}

	return nil
}

// cleanNumberString removes the thousand separators and the trailing unit from
// a number as it is written in a CSV file
func cleanNumberString(s string) string {
	s = strings.Replace(s, ".", "", -1) //remove possible dots
	return strings.Trim(s, "%\n ")
}

// isTrueString reports whether s is one of the spellings of "true" the import
// accepts. Everything else is false
func isTrueString(s string) bool {
	switch strings.ToLower(s) {
	case "wahr", "true", "ja", "yes", "1", "x":
		return true
	}

	return false
}

// GetExcelLineValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
func (ii *Importer[C]) GetExcelLineValues(line int, row *Row) *reflect.Value {
	v := reflect.New(ii.structType).Elem()
	var err error

	for excelRef, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		col := value.PosInData
		if col < len(row.Cells) {
			cell := row.Cells[col]

			_, err = ii.getCellValue(cell, f, &v)
			if err != nil {
				ii.errorList.AddErrorString("error on getting field values for:%v error: %v", excelRef, err.Error())
			}
		}
	}

	return &v
}

// GetExcelLineColValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
func (ii *Importer[C]) GetExcelLineColValues(line int, row *Row) *reflect.Value {
	v := reflect.New(ii.structType).Elem()
	var err error

	for excelRef, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		col := int(excel_cols.GetPosFromString(value.Col))

		if col < len(row.Cells) {
			cell := row.Cells[col]

			_, err = ii.getCellValue(cell, f, &v)
			if err != nil {
				ii.errorList.AddErrorString("error on getting field values for:%v error: %v", excelRef, err.Error())
			}
		}
	}

	return &v
}

// GetExcelPosValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
func (ii *Importer[C]) GetExcelPosValues(sh *Sheet) *reflect.Value {
	v := reflect.New(ii.structType).Elem()

	for excelRef, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		colNr, rowNr, err := cellRefToPos(value.ExcelPos)
		if err != nil {
			ii.errorList.AddErrorString("error on getting excel cooridante for: %s error: %v", value.ExcelPos, err.Error())
		}

		cell := sh.Cell(rowNr, colNr)

		_, err = ii.getCellValue(cell, f, &v)
		if err != nil {
			ii.errorList.AddErrorString("error on getting field values for:%v error: %v", excelRef, err.Error())
		}
	}

	return &v
}

// ValidateStruct validates the struct data after it has been read
func (ii *Importer[C]) ValidateStruct(line int, data interface{}) {
	//validate the data
	errVal := validate.Struct(data)
	if errVal != nil {
		col := 0
		excelRef := ""

		for _, err := range errVal.(validator.ValidationErrors) {
			if field := ii.fieldForName(err.Field()); field != nil {
				if ii.Config.dataType == ExcelLine {
					excelRef = ii.dataHeaders[field.PosInData]
					col = field.PosInData
				} else {
					excelRef = field.ExcelPos
					col = -1
				}
			}

			err := fmt.Errorf("error validating field: '%s' tag: '%s' value: '%v' param:'%s'", err.Field(), err.ActualTag(), err.Value(), err.Param())
			ii.errorList.AddValidation(line+1, col, excelRef, err)
		}
	}
}

func (ii *Importer[C]) fieldForName(fieldName string) *field {
	for _, f := range ii.fields {
		if f.Field.Name == fieldName {
			return f
		}
	}
	return nil
}

// getCellValue converts the value of an Excel cell into the struct field it
// belongs to. The bool reports whether a value was assigned, which is what
// tells a nullable field apart from one that stays unset
func (ii *Importer[C]) getCellValue(cell *Cell, f reflect.Value, v *reflect.Value) (bool, error) {
	kk := f.Kind().String()

	if cell.Value == "" || slices.Contains(ii.Config.EmptyValueStrings, cell.Value) {
		return false, nil
	}

	switch f.Interface().(type) {
	case int, int32, int64:
		i, err := cell.Int()
		if err != nil {
			return false, fmt.Errorf("getting reflect on field: %s int: %w", kk, err)
		}
		f.SetInt(int64(i))

	case uint, uint32, uint64:
		i, err := cell.Int()
		if err != nil {
			return false, fmt.Errorf("getting reflect on field: %s int: %w", kk, err)
		}
		f.SetUint(uint64(i))

	case string:
		f.SetString(cell.String())

	case time.Time:
		//a date is a number in Excel, anything else is left alone
		if cell.Type() != CellTypeNumeric {
			return false, nil
		}

		t, err := cell.GetTime()
		if err != nil {
			return false, fmt.Errorf("getting reflect on field: %s time.Time: %w", kk, err)
		}
		f.Set(reflect.ValueOf(t))

	case float32, float64:
		i, err := cell.Float()
		if err != nil {
			return false, fmt.Errorf("getting reflect on field: %s float: %w", kk, err)
		}
		f.SetFloat(i)

	case bool:
		f.SetBool(isTrueString(cell.String()))

	default:
		inner, ok := nullableValue(f)
		if !ok {
			return false, fmt.Errorf("undefined kind '%v'", f.Kind())
		}

		set, err := ii.getCellValue(cell, inner, v)
		if err != nil || !set {
			return false, err
		}
		setNullableValid(f)
	}

	return true, nil
}
