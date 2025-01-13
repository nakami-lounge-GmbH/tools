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

var (
	unidocLicenseKey   = os.Getenv("UNIDOC_LICENSE_KEY")
	unidocCustomerName = os.Getenv("UNIDOC_CUSTOMER_NAME")
)

func initExcelLic() {
	if err := license.SetLicenseKey(unidocLicenseKey, unidocCustomerName); err != nil {
		log.Println("Error loading unidocPDFLicense:", err)
		os.Exit(-1)
	}

	if err := ll.SetLicenseKey(unidocLicenseKey, unidocCustomerName); err != nil {
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
	el := new(ErrorList)
	i, err := NewExcelLineImporter[data](&ExcelLineConfig{
		SheetName:         "Daten",
		SheetNumber:       0,
		OffsetRow:         3,
		OffsetCol:         2,
		FileBytes:         fileBytes,
		LineCountToRead:   1,
		EmptyValueStrings: nil,
	}, el)

	if err != nil {
		if el.HasAny() {
			log.Println("Errors:", el.String())
		}
		log.Fatalln("With error:", err)
	}

	if el.HasAny() {
		log.Fatalln(el.String())
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
	el := new(ErrorList)
	i, err := NewExcelPageImporter[data](&ExcelPageConfig{
		FileBytes:         fileBytes,
		SheetName:         "Daten",
		SheetNumber:       0,
		EmptyValueStrings: []string{},
	}, el)

	if err != nil {
		if el.HasAny() {
			log.Println("Errors:", el.String())
		}
		log.Fatalln("With error:", err)
	}

	if el.HasAny() {
		log.Fatalln(el.String())
	}

	fmt.Println("data:", len(i.Data), i.Data)
}
