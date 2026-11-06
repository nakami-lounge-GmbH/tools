package importer

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func readFixture(t *testing.T) []byte {
	t.Helper()

	fileBytes, err := os.ReadFile("test1.xlsx")
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	return fileBytes
}

func TestNewExcelLineImporter(t *testing.T) {
	type data struct {
		Name       string    `header:"Name"`
		Vorname    string    `header:"Vorname"`
		Geburtstag time.Time `header:"Geburtstag"`
		CntKids    int       `header:"Anzahl Kinder"`
		Test       float64   `header:"Test"`
	}

	el := new(ErrorList)
	i, err := NewExcelLineImporter[data](&ExcelLineConfig{
		Config{
			SheetName:         "Daten",
			SheetNumber:       0,
			OffsetRow:         3,
			FileBytes:         readFixture(t),
			LineCountToRead:   1,
			EmptyValueStrings: nil,
		},
	}, el)

	if err != nil {
		t.Fatalf("with error: %v errorlist: %s", err, el.String())
	}
	if el.HasAny() {
		t.Fatal(el.String())
	}

	want := []data{
		{Name: "Polder", Vorname: "Udo", Geburtstag: time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC), CntKids: 4, Test: 0},
	}
	if !reflect.DeepEqual(i.Data, want) {
		t.Errorf("data mismatch\n got: %v\nwant: %v", i.Data, want)
	}
}

func TestNewExcelPageImporter(t *testing.T) {
	type data struct {
		Name       string    `excel_pos:"B4"`
		Vorname    string    `excel_pos:"C5"`
		Geburtstag time.Time `excel_pos:"E4"`
		CntKids    int       `excel_pos:"F4"`
		Test       float64   `excel_pos:"H5"`
	}

	el := new(ErrorList)
	i, err := NewExcelPageImporter[data](&ExcelPageConfig{
		Config{
			FileBytes:         readFixture(t),
			SheetName:         "Daten",
			SheetNumber:       0,
			EmptyValueStrings: []string{},
		},
	}, el)

	if err != nil {
		t.Fatalf("with error: %v errorlist: %s", err, el.String())
	}
	if el.HasAny() {
		t.Fatal(el.String())
	}

	want := []data{
		{Name: "Polder", Vorname: "Michaela", Geburtstag: time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC), CntKids: 4, Test: 777.7},
	}
	if !reflect.DeepEqual(i.Data, want) {
		t.Errorf("data mismatch\n got: %v\nwant: %v", i.Data, want)
	}
}

func TestNewExcelLineColImporter(t *testing.T) {
	type data struct {
		Name       string    `col:"B"`
		Vorname    string    `col:"C"`
		Geburtstag time.Time `col:"E"`
		CntKids    int       `col:"F"`
		Test       float64   `col:"H"`
	}

	el := new(ErrorList)
	i, err := NewExcelLineColImporter[data](&ExcelLineColConfig{
		Config{
			SheetName:         "Daten",
			SheetNumber:       0,
			OffsetRow:         3,
			FileBytes:         readFixture(t),
			LineCountToRead:   5,
			EmptyValueStrings: nil,
		},
	}, el)

	if err != nil {
		t.Fatalf("with error: %v errorlist: %s", err, el.String())
	}
	if el.HasAny() {
		t.Fatal(el.String())
	}

	want := []data{
		{Name: "Polder", Vorname: "Udo", Geburtstag: time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC), CntKids: 4, Test: 0},
		{Name: "Rauna", Vorname: "Michaela", Geburtstag: time.Date(1998, 3, 25, 0, 0, 0, 0, time.UTC), CntKids: 2, Test: 777.7},
	}
	if !reflect.DeepEqual(i.Data, want) {
		t.Errorf("data mismatch\n got: %v\nwant: %v", i.Data, want)
	}
}

// TestGetSheet covers how the sheet is picked from the config
func TestGetSheet(t *testing.T) {
	fileBytes := readFixture(t)

	tests := []struct {
		name        string
		sheetName   string
		sheetNumber int
		wantSheet   string
		wantErr     bool
	}{
		{name: "by name", sheetName: "Daten", wantSheet: "Daten"},
		{name: "by number", sheetNumber: 1, wantSheet: "Daten"},
		{name: "unknown name", sheetName: "gibtsnicht", wantErr: true},
		{name: "number out of range", sheetNumber: 2, wantErr: true},
		{name: "nothing set", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ii := Importer[struct{}]{Config: &Config{
				SheetName:   tt.sheetName,
				SheetNumber: tt.sheetNumber,
				FileBytes:   fileBytes,
			}}

			sh, err := ii.GetSheet()
			if sh != nil {
				defer func() { _ = sh.Close() }()
			}

			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got sheet %v", sh)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sh == nil {
				t.Fatal("no sheet and no error")
			}
			if sh.Name != tt.wantSheet {
				t.Errorf("sheet name: got %q want %q", sh.Name, tt.wantSheet)
			}
			//the fixture holds its header in row 3 and data up to row 5
			if len(sh.Rows) != 5 {
				t.Errorf("row count: got %d want 5", len(sh.Rows))
			}
		})
	}
}

// TestCellValues covers the raw and the displayed value as well as the type of
// a cell, which is what the conversion into the struct fields builds on
func TestCellValues(t *testing.T) {
	ii := Importer[struct{}]{Config: &Config{SheetName: "Daten", FileBytes: readFixture(t)}}

	sh, err := ii.GetSheet()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = sh.Close() }()

	tests := []struct {
		ref      string
		row      int
		col      int
		value    string
		cellType CellType
	}{
		{ref: "B3", row: 2, col: 1, value: "Name", cellType: CellTypeString},
		{ref: "E4", row: 3, col: 4, value: "27669", cellType: CellTypeNumeric},
		{ref: "F4", row: 3, col: 5, value: "4", cellType: CellTypeNumeric},
		{ref: "H5", row: 4, col: 7, value: "777.7", cellType: CellTypeNumeric},
		{ref: "H4 (not in the file)", row: 3, col: 7, value: "", cellType: CellTypeEmpty},
		{ref: "ZZ99 (outside the data)", row: 98, col: 701, value: "", cellType: CellTypeEmpty},
	}

	for _, tt := range tests {
		cell := sh.Cell(tt.row, tt.col)
		if cell == nil {
			t.Fatalf("%s: no cell", tt.ref)
		}
		if cell.Value != tt.value {
			t.Errorf("%s: value got %q want %q", tt.ref, cell.Value, tt.value)
		}
		if cell.Type() != tt.cellType {
			t.Errorf("%s: type got %v want %v", tt.ref, cell.Type(), tt.cellType)
		}
	}

	//the date is stored as a number and displayed with its number format
	date := sh.Cell(3, 4)
	got, err := date.GetTime()
	if err != nil {
		t.Fatalf("E4 GetTime: %v", err)
	}
	if want := time.Date(1975, 10, 2, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("E4 as time: got %v want %v", got, want)
	}
	if date.String() == date.Value {
		t.Errorf("E4: expected the displayed value to differ from the raw value %q", date.Value)
	}
}
