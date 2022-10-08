package uploader

import "strings"

var (
	FileTypePDF   = AllowedFileType{Name: "pdf", Type: "application/pdf"}
	FileTypePNG   = AllowedFileType{Name: "png", Type: "image/png"}
	FileTypeJPG   = AllowedFileType{Name: "jpg", Type: "image/jpeg"}
	FileTypeExcel = AllowedFileType{Name: "xlsx", Type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}
	FileTypeCSV   = AllowedFileType{Name: "csv", Type: "text/csv"}
)

type AllowedFileType struct {
	Name string
	Type string
}

type AllowedFileTypes []AllowedFileType

func NewAllowedFileTypeChecker(allowedTypes []AllowedFileType) AllowedFileTypes {
	var ret AllowedFileTypes
	ret = allowedTypes
	return ret
}

func (m AllowedFileTypes) IsAllowed(typeToCheck string) bool {
	for _, t := range m {
		if t.Type == typeToCheck {
			return true
		}
	}
	return false
}

func (m AllowedFileTypes) GetTypesString() string {
	t := make([]string, len(m))
	for i, v := range m {
		t[i] = v.Name
	}
	return strings.Join(t, ", ")
}
