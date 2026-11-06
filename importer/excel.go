package importer

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// CellType describes the kind of value a cell holds. It is the subset of the
// Excel cell types the importer needs to tell apart.
type CellType int

const (
	//CellTypeEmpty is reported for cells without any content
	CellTypeEmpty CellType = iota

	//CellTypeString is reported for text cells
	CellTypeString

	//CellTypeNumeric is reported for numbers. Dates and times are numbers in
	//Excel, so date cells are numeric as well
	CellTypeNumeric

	//CellTypeBool is reported for boolean cells
	CellTypeBool

	//CellTypeError is reported for cells holding an Excel error like #N/A
	CellTypeError
)

// Cell is a single cell of a Sheet
type Cell struct {
	//Value is the value as it is stored in the file: the plain number for
	//numbers and dates, the text itself for text cells
	Value string

	//formatted is the value as Excel displays it, with the number format of the
	//cell applied
	formatted string

	sheet *Sheet

	//row and col are the 1-based coordinates of the cell
	row int
	col int

	cellType    CellType
	cellTypeSet bool
}

// String returns the value as Excel displays it, with the number format of the
// cell applied
func (c *Cell) String() string {
	if c == nil {
		return ""
	}
	return c.formatted
}

// Type returns the type of the value in the cell
func (c *Cell) Type() CellType {
	if c == nil {
		return CellTypeEmpty
	}
	if !c.cellTypeSet {
		c.cellType = c.resolveType()
		c.cellTypeSet = true
	}
	return c.cellType
}

// Int returns the value as int
func (c *Cell) Int() (int, error) {
	if c == nil {
		return 0, fmt.Errorf("no cell")
	}
	return strconv.Atoi(c.Value)
}

// Int64 returns the value as int64
func (c *Cell) Int64() (int64, error) {
	if c == nil {
		return 0, fmt.Errorf("no cell")
	}
	return strconv.ParseInt(c.Value, 10, 64)
}

// Float returns the value as float64
func (c *Cell) Float() (float64, error) {
	if c == nil {
		return 0, fmt.Errorf("no cell")
	}
	return strconv.ParseFloat(c.Value, 64)
}

// GetTime reads the value as date/time. Excel stores both as a number of days
// since an epoch, which is taken from the workbook the cell belongs to
func (c *Cell) GetTime() (time.Time, error) {
	if c == nil {
		return time.Time{}, fmt.Errorf("no cell")
	}

	v, err := c.Float()
	if err != nil {
		return time.Time{}, err
	}

	date1904 := false
	if c.sheet != nil {
		date1904 = c.sheet.date1904
	}

	return excelize.ExcelDateToTime(v, date1904)
}

// resolveType asks the file for the type of the cell. Numbers carry no type of
// their own in the file and come back unset, so those are classified by their
// value
func (c *Cell) resolveType() CellType {
	if c.Value == "" {
		return CellTypeEmpty
	}

	if c.sheet != nil && c.sheet.file != nil {
		if ref, err := excelize.CoordinatesToCellName(c.col, c.row); err == nil {
			t, err := c.sheet.file.GetCellType(c.sheet.Name, ref)
			if err == nil {
				switch t {
				case excelize.CellTypeSharedString, excelize.CellTypeInlineString:
					return CellTypeString
				case excelize.CellTypeBool:
					return CellTypeBool
				case excelize.CellTypeError:
					return CellTypeError
				case excelize.CellTypeDate, excelize.CellTypeNumber:
					return CellTypeNumeric
				}
			}
		}
	}

	if _, err := strconv.ParseFloat(c.Value, 64); err == nil {
		return CellTypeNumeric
	}

	return CellTypeString
}

// Row is a single row of a Sheet
type Row struct {
	//Cells holds the cells of the row, indexed from 0 for column "A". Empty
	//cells at the end of the row are not part of the slice
	Cells []*Cell
}

// Sheet is one worksheet of an Excel file, read into memory
type Sheet struct {
	//Name is the name of the sheet inside the file
	Name string

	//Rows holds the rows of the sheet, indexed from 0 for the first row. Rows
	//without content are part of the slice but hold no cells
	Rows []*Row

	file     *excelize.File
	date1904 bool
}

// Cell returns the cell at the given 0-based row and column. Coordinates
// outside the data of the sheet return an empty cell, not nil
func (sh *Sheet) Cell(row, col int) *Cell {
	if sh != nil && row >= 0 && row < len(sh.Rows) {
		cells := sh.Rows[row].Cells
		if col >= 0 && col < len(cells) {
			return cells[col]
		}
	}

	return &Cell{sheet: sh, row: row + 1, col: col + 1}
}

// Close releases the file the sheet was read from. It must be called once the
// sheet is no longer used
func (sh *Sheet) Close() error {
	if sh == nil || sh.file == nil {
		return nil
	}
	return sh.file.Close()
}

// openSheet reads one sheet of an Excel file. The sheet is taken by name, or,
// if no name is given, by its 1-based number
func openSheet(fileBytes []byte, sheetName string, sheetNumber int) (*Sheet, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}

	name := sheetName
	if name == "" {
		sheets := f.GetSheetList()
		if sheetNumber < 1 || sheetNumber > len(sheets) {
			_ = f.Close()
			return nil, fmt.Errorf("error reading sheet index %d", sheetNumber)
		}
		name = sheets[sheetNumber-1]
	} else if idx, err := f.GetSheetIndex(name); err != nil || idx < 0 {
		_ = f.Close()
		return nil, fmt.Errorf("error reading sheet %s", sheetName)
	}

	//the values as stored in the file, used for all typed conversions
	raw, err := f.GetRows(name, excelize.Options{RawCellValue: true})
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	//the same values as Excel displays them, used wherever the importer reads a
	//cell as text
	formatted, err := f.GetRows(name)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	sh := &Sheet{Name: name, file: f}

	if props, err := f.GetWorkbookProps(); err == nil && props.Date1904 != nil {
		sh.date1904 = *props.Date1904
	}

	sh.Rows = make([]*Row, len(raw))
	for r, rawRow := range raw {
		row := &Row{Cells: make([]*Cell, len(rawRow))}

		for c, v := range rawRow {
			cell := &Cell{Value: v, formatted: v, sheet: sh, row: r + 1, col: c + 1}
			if r < len(formatted) && c < len(formatted[r]) {
				cell.formatted = formatted[r][c]
			}
			row.Cells[c] = cell
		}

		sh.Rows[r] = row
	}

	return sh, nil
}

// cellRefToPos returns the 0-based column and row of an Excel cell reference
// like "B4"
func cellRefToPos(ref string) (col int, row int, err error) {
	c, r, err := excelize.CellNameToCoordinates(ref)
	if err != nil {
		return 0, 0, err
	}
	return c - 1, r - 1, nil
}
