package importer

import (
	"github.com/tealeg/xlsx"
	"golang.org/x/exp/slices"
	"strings"
)

func headerToStrings(row *xlsx.Row) ([]string, int) {
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

func rowToStrings(row *xlsx.Row, length int) []string {
	var ret []string
	for i := 0; i < length; i++ {
		if len(row.Cells) > i {
			c := row.Cells[i]
			if c.Type() == xlsx.CellTypeNumeric && !slices.Contains([]string{"", "general", "@"}, c.NumFmt) {
				form := c.NumFmt

				t, err := c.GetTime(false)
				if err == nil {
					if strings.Contains(form, "h") || strings.Contains(form, "h") {
						ret = append(ret, t.Format("02/01/2006 15:04:05"))
					} else {
						ret = append(ret, t.Format("02/01/2006"))
					}
				} else {
					ret = append(ret, "")
				}
			} else {
				ret = append(ret, c.String())
			}
		} else {
			ret = append(ret, "")
		}
	}

	return ret
}
