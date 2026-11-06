package importer

import (
	"strings"
)

func headerToStrings(row *Row) ([]string, int) {
	var ret []string
	for _, c := range row.Cells {
		ret = append(ret, c.String())
	}
	return ret, len(ret)
}

func isEmpty(line []string) bool {
	for _, s := range line {
		if strings.Trim(s, " \n\t\r") != "" {
			return false
		}
	}
	return true
}

// rowToStrings returns the cells of the row as text, as Excel displays them,
// padded to length. The result is used to tell empty rows apart from rows with
// content, the typed values are read from the cells themselves
func rowToStrings(row *Row, length int) []string {
	ret := make([]string, 0, length)
	for i := 0; i < length; i++ {
		if i < len(row.Cells) {
			ret = append(ret, row.Cells[i].String())
		} else {
			ret = append(ret, "")
		}
	}

	return ret
}
