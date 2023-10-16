package importer

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/tealeg/xlsx"
	"github.com/volatiletech/null/v8"
	"golang.org/x/exp/slices"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ImportType int

const (
	ExcelLine ImportType = iota
	ExcelPage ImportType = iota
)

const (
	tagNameLine       = "header"
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

func (ii *Importer[C]) GetSheet() (*xlsx.Sheet, error) {
	xlsFile, err := xlsx.OpenBinary(ii.Config.FileBytes)
	if err != nil {
		return nil, err
	}

	var sh *xlsx.Sheet
	var ok bool

	if ii.Config.SheetName != "" {
		sh, ok = xlsFile.Sheet[ii.Config.SheetName]
		if !ok {
			return nil, fmt.Errorf("error reading sheet %s", ii.Config.SheetName)
		}
	} else if ii.Config.SheetNumber != 0 {
		if ii.Config.SheetNumber <= len(xlsFile.Sheets) {
			sh = xlsFile.Sheets[ii.Config.SheetNumber-1]
		} else {
			return nil, fmt.Errorf("error reading sheet index %d", ii.Config.SheetNumber)
		}
	}

	return sh, nil
}

type field struct {
	PosInData int    //used in line import. Specifies the position in the data
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

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tagName := tagNameLine
		if ii.Config.dataType == ExcelPage {
			tagName = tagNamePage
		}

		tag := f.Tag.Get(tagName)
		if tag != "" {
			if ii.Config.dataType == ExcelLine {
				ii.fields[strings.ToLower(tag)] = &field{Field: f, PosInData: -1}
			} else if ii.Config.dataType == ExcelPage {
				ii.fields[strings.ToLower(tag)] = &field{Field: f, ExcelPos: tag}
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
			fieldVal := f.Interface()
			switch fieldVal.(type) {
			case int, int32, int64:
				s := strings.Replace(data[col], ".", "", -1) //remove possible dots
				s = strings.Trim(s, "%\n ")
				if i, err := strconv.Atoi(s); err == nil {
					f.SetInt(int64(i))
				} else {
					ii.errorList.AddValidation(line, col, header, err)
					return nil
				}

			case string:
				f.SetString(data[col])

			case time.Time:
				//check if format-tag set
				format := value.Field.Tag.Get(tagNameDateFormat)
				if format == "" {
					ii.errorList.AddValidation(line, col, header, fmt.Errorf("No format tag on structfield '%s'", header))
					return nil
				}
				t, err := time.Parse(format, data[col])
				if err != nil {
					ii.errorList.AddValidation(line, col, header, fmt.Errorf("Could not parse value '%s' as date-time with format '%s'", data[col], format))
					return nil
				}
				f.Set(reflect.ValueOf(t))

			case float32, float64:
				s := data[col]
				s = strings.Replace(s, ",", ".", 1) //we need to have only one format
				if i, err := strconv.ParseFloat(s, 64); err == nil {
					f.SetFloat(float64(i))
				} else {
					ii.errorList.AddValidation(line, col, header, err)
					return nil
				}

			case bool:
				s := strings.ToLower(data[col])
				if s == "wahr" || s == "true" || s == "ja" || s == "yes" || s == "1" || s == "x" {
					f.SetBool(true)
				} else {
					f.SetBool(false)
				}
			case null.String:
				f.Set(reflect.ValueOf(null.StringFrom(data[col])))
			case null.Time:
				//check if format-tag set
				format := value.Field.Tag.Get(tagNameDateFormat)
				if format == "" {
					ii.errorList.AddValidation(line, col, header, fmt.Errorf("No format tag on structfield '%s'", header))
					return nil
				}
				t, err := time.Parse(format, data[col])
				if err != nil {
					ii.errorList.AddValidation(line, col, header, fmt.Errorf("Could not parse value '%s' as date-time with format '%s'", data[col], format))
					return nil
				}
				f.Set(reflect.ValueOf(null.TimeFrom(t)))
			case null.Int:

				s := strings.Replace(data[col], ".", "", -1) //remove possible dots
				s = strings.Trim(s, "%\n ")

				if i, err := strconv.Atoi(s); err == nil {
					f.Set(reflect.ValueOf(null.IntFrom(i)))
				} else {
					ii.errorList.AddValidation(line, col, header, err)
					return nil
				}
			case null.Uint:

				s := strings.Replace(data[col], ".", "", -1) //remove possible dots
				s = strings.Trim(s, "%\n ")

				if i, err := strconv.ParseUint(s, 10, 32); err == nil {
					f.Set(reflect.ValueOf(null.UintFrom(uint(i))))
				} else {
					ii.errorList.AddValidation(line, col, header, err)
					return nil
				}

			case null.Bool:
				s := strings.ToLower(data[col])
				if s == "wahr" || s == "true" || s == "ja" || s == "yes" || s == "1" || s == "x" {
					f.Set(reflect.ValueOf(null.BoolFrom(true)))
				} else {
					f.Set(reflect.ValueOf(null.BoolFrom(false)))
				}

			default:
				ii.errorList.AddValidation(line, col, header, fmt.Errorf("undefined kind '%v' on structfield '%s'", f.Kind(), header))
				return nil
			}
		}
	}

	return &v
}

// GetExcelLineValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
func (ii *Importer[C]) GetExcelLineValues(line int, row *xlsx.Row) *reflect.Value {
	v := reflect.New(ii.structType).Elem()
	var err error

	for excelRef, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		col := value.PosInData
		if col<len(row.Cells){
			cell := row.Cells[col]

			err = ii.getCellValue(cell, f, &v)
			if err != nil {
				ii.errorList.AddErrorString("error on getting field values for:%v error: %v", excelRef, err.Error())
			}
		}
	}

	return &v
}

// GetExcelPosValues returns an instance of the array element.
// It must be cast to the concrete type by the caller
func (ii *Importer[C]) GetExcelPosValues(sh *xlsx.Sheet) *reflect.Value {
	v := reflect.New(ii.structType).Elem()

	for excelRef, value := range ii.fields {
		f := v.FieldByName(value.Field.Name)
		colNr, rowNr, err := xlsx.GetCoordsFromCellIDString(value.ExcelPos)
		if err != nil {
			ii.errorList.AddErrorString("error on getting excel cooridante for: %s error: %v", value.ExcelPos, err.Error())
		}

		cell := sh.Cell(rowNr, colNr)

		err = ii.getCellValue(cell, f, &v)
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

func (ii *Importer[C]) getCellValue(cell *xlsx.Cell, f reflect.Value, v *reflect.Value) error {
	//v := reflect.New(ii.structType).Elem()
	kk := f.Kind().String()
	if cell.Value != "" && !slices.Contains(ii.Config.EmptyValueStrings, cell.Value) {
		fieldVal := f.Interface()
		switch fieldVal.(type) {
		case int, int32, int64:
			i, err := cell.Int()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s int: %w", kk, err)
			}
			f.SetInt(int64(i))
		case uint, uint32, uint64:
			i, err := cell.Int()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s int: %w", kk, err)
			}
			f.SetUint(uint64(i))
		case string:
			f.SetString(cell.String())
		case time.Time:
			t := cell.Type()
			if t == xlsx.CellTypeNumeric {
				t, err := cell.GetTime(false)
				if err != nil {
					return fmt.Errorf("getting reflect on field: %s time.Time: %w", kk, err)
				}
				f.Set(reflect.ValueOf(t))
			}
		case float32, float64:
			i, err := cell.Float()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s float: %w", kk, err)
			}

			f.SetFloat(i)
		case bool:
			s := strings.ToLower(cell.String())
			if s == "wahr" || s == "true" || s == "ja" || s == "yes" || s == "1" || s == "x" {
				f.SetBool(true)
			} else {
				f.SetBool(false)
			}
		case null.String:
			f.Set(reflect.ValueOf(null.StringFrom(cell.String())))
		case null.Time:
			if cell.Type() == xlsx.CellTypeNumeric {
				t, err := cell.GetTime(false)
				if err != nil {
					return fmt.Errorf("getting reflect on field: %s null.Time: %w", kk, err)
				}
				f.Set(reflect.ValueOf(null.TimeFrom(t)))
			}
		case null.Float32:
			i, err := cell.Float()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s nullFloat32: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.Float32From(float32(i))))
		case null.Float64:
			i, err := cell.Float()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s nullFloat64: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.Float64From(i)))
		case null.Int:
			i, err := cell.Int()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s null.Int: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.IntFrom(i)))
		case null.Int64:
			i, err := cell.Int64()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s null.Int64: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.Int64From(i)))
		case null.Int32:
			i, err := cell.Int()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s null.Int32: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.Int32From(int32(i))))
		case null.Uint:
			i, err := cell.Int()
			if err != nil {
				return fmt.Errorf("getting reflect on field: %s null.UInt: %w", kk, err)
			}
			f.Set(reflect.ValueOf(null.UintFrom(uint(i))))
		case null.Bool:
			s := strings.ToLower(cell.String())
			if s == "wahr" || s == "true" || s == "ja" || s == "yes" || s == "1" || s == "x" {
				f.Set(reflect.ValueOf(null.BoolFrom(true)))
			} else {
				f.Set(reflect.ValueOf(null.BoolFrom(false)))
			}
		default:
			return fmt.Errorf("undefined kind '%v'", f.Kind())
		}
	}

	return nil
}
