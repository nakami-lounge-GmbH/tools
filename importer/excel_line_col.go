package importer

import (
	"fmt"
)

type ExcelLineColImporter[C any] struct {
	Importer[C]

	LinesProcessed int
	LinesImported  int
}

type ExcelLineColConfig struct {
	Config
}

// NewExcelLineColImporter returns a new line-importer, where the cols are specified via "col" tag
// the importer still needs a header line but only for determining the number of cols
// the data tag is col "A", "B", "AA" ...
// Offset must be set to the line numbers before the header line or 0
func NewExcelLineColImporter[C any](config *ExcelLineColConfig, eL *ErrorList) (ExcelLineImporter[C], error) {
	var r ExcelLineImporter[C]

	r.Importer = newImorter[C](&Config{
		dataType:          ExcelLineCol,
		SheetName:         config.SheetName,
		SheetNumber:       config.SheetNumber,
		OffsetRow:         config.OffsetRow,
		FileBytes:         config.FileBytes,
		LineCountToRead:   config.LineCountToRead,
		EmptyValueStrings: config.EmptyValueStrings,
	}, eL)

	var err error

	sh, err := r.GetSheet()
	if err != nil {
		return r, fmt.Errorf("error reading sheet %w", err)
	}

	headerCount := 0
	var line []string
	linesProcessed := 0

	for l, row := range sh.Rows {
		if l == r.Importer.Config.OffsetRow-1 || (l == 0 && r.Importer.Config.OffsetRow == 0) {
			headerCount = len(row.Cells)
		} else if l >= r.Config.OffsetRow {
			if r.Config.LineCountToRead != 0 && linesProcessed >= r.Config.LineCountToRead {
				break //exit for loop if number of lines reached
			}

			line = rowToStrings(row, headerCount)

			if !isEmpty(line) {
				r.LinesProcessed++
				mV := r.GetExcelLineColValues(l, row)
				if eL.HasErrors() {
					return r, nil
				}
				if mV != nil {
					m := mV.Interface().(C)
					r.Data = append(r.Data, m)
					r.ValidateStruct(l, m)
				}
			}
			linesProcessed++
		}
	}

	return r, nil
}
