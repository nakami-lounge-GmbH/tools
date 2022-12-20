package importer

import (
	"fmt"
	ll "github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unipdf/v3/common/license"
	"log"
	"os"
	"testing"
	"time"
)

const (
	unidocLicenseKey = `-----BEGIN UNIDOC LICENSE KEY-----
eyJsaWNlbnNlX2lkIjoiYzcwM2QxNDUtYWMyYi00NTkxLTYwYzItNmY1ZTk4YmQ4Y2QyIiwiY3VzdG9tZXJfaWQiOiI5ZWEyZDUyMS01YmZkLTRmMjItNzI2YS04MzlmODcyMjcyMjYiLCJjdXN0b21lcl9uYW1lIjoibmFrYW1pIGxvdW5nZSBHbWJIIiwiY3VzdG9tZXJfZW1haWwiOiJtZkBuYWthbWkuZGUiLCJ0aWVyIjoiYnVzaW5lc3MiLCJjcmVhdGVkX2F0IjoxNjQ3ODgwNzAzLCJleHBpcmVzX2F0IjoxNjY3MDAxNTk5LCJjcmVhdG9yX25hbWUiOiJVbmlEb2MgU3VwcG9ydCIsImNyZWF0b3JfZW1haWwiOiJzdXBwb3J0QHVuaWRvYy5pbyIsInVuaXBkZiI6dHJ1ZSwidW5pb2ZmaWNlIjp0cnVlLCJ0cmlhbCI6ZmFsc2V9
+
OpHtrSR01n718vlRFQoAG+LkRxsNdz6Xmzqx3769D0mM3z5W3e7WGQyeWQySaGC4KAcwDcW2dTEpLjiFgZRwm9uKJ4Pz1Ro6TNqozwTjn9uGE6fOf4bI7z/15EHkri+oqteHelnyRJuuA5dwGMNbp9Q0aBu22J/WM0M7W6ktW12k3m3cXtoXbZ7LfTsw3x61Ep2ekG9w9lqmKwZ+8AcfD4IxukL5j70MswejWnalKHQcqpu+xgOeMtMdhYpsqwn9jokxFOeY+Owyh08BRudBD9MlQYETkOm0//xSBhuX+MdWxnYEOyV1JBAm+j/ZRKc2LBCvdXA5qa/Pe7PrarNrVA==
-----END UNIDOC LICENSE KEY-----`
)

func initExcelLic() {
	if err := license.SetLicenseKey(unidocLicenseKey, "Nakami Lounge GmbH"); err != nil {
		log.Println("Error loading unidocPDFLicense:", err)
		os.Exit(-1)
	}

	if err := ll.SetLicenseKey(unidocLicenseKey, "nakami lounge GmbH"); err != nil {
		log.Println("Error loading unidocOfficeLicense:", err)
		os.Exit(-1)
	}
}

func TestNewExcelLineImporter(t *testing.T) {
	initExcelLic()

	type data struct {
		Name       string    `header:"Name"`
		Vorname    string    `header:"Vorname"`
		Geburtstag time.Time `header:"Geburtstag"`
		CntKids    int       `header:"Anzahl Kinder"`
		Test       float64   `header:"Test"`
	}

	fileBytes, err := os.ReadFile("test1.xlsx")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	//var d []data
	i, err := NewExcelLineImporter[data](&ExcelLineConfig{
		FileBytes:       fileBytes,
		SheetName:       "Daten",
		OffsetRow:       3,
		OffsetCol:       2,
		LineCountToRead: 1,
	})

	if err != nil {
		if i.ErrorList.HasAny() {
			log.Println("Errors:", i.ErrorList.String())
		}
		log.Fatalln("With error:", err)
	}

	if i.ErrorList.HasAny() {
		log.Fatalln(i.ErrorList.String())
	}

	fmt.Println("data:", len(i.Data), i.Data)
}

func TestNewExcelPageImporter(t *testing.T) {
	initExcelLic()

	type data struct {
		Name       string    `excel_pos:"B4"`
		Vorname    string    `excel_pos:"C5"`
		Geburtstag time.Time `excel_pos:"E4"`
		CntKids    int       `excel_pos:"F4"`
		Test       float64   `excel_pos:"H5"`
	}

	fileBytes, err := os.ReadFile("test1.xlsx")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	//var d []data
	i, err := NewExcelPageImporter[data](&ExcelPageConfig{
		FileBytes: fileBytes,
		SheetName: "Daten",
	})

	if err != nil {
		if i.ErrorList.HasAny() {
			log.Println("Errors:", i.ErrorList.String())
		}
		log.Fatalln("With error:", err)
	}

	if i.ErrorList.HasAny() {
		log.Fatalln(i.ErrorList.String())
	}

	fmt.Println("data:", len(i.Data), i.Data)
}
