package importer

type ExcelPageImporter[C any] struct {
	Importer[C]
}

type ExcelPageConfig struct {
	SheetName         string   //either name or number (0-indexed)
	SheetNumber       int      //either name or number (0-indexed)
	FileBytes         []byte   //data from the excel file
	EmptyValueStrings []string //specifies values that should be treated as empty
}

func NewExcelPageImporter[C any](config *ExcelPageConfig, eL *ErrorList) (ExcelPageImporter[C], error) {
	var r ExcelPageImporter[C]

	r.Importer = newImorter[C](&Config{
		dataType:          ExcelPage,
		SheetName:         config.SheetName,
		SheetNumber:       config.SheetNumber,
		FileBytes:         config.FileBytes,
		EmptyValueStrings: config.EmptyValueStrings,
	}, eL)

	var err error

	sh, err := r.GetSheet()
	if err != nil {
		eL.AddErrorMsg(err, "reading sheet")
		return r, nil
	}

	mV := r.GetExcelPosValues(sh)
	if eL.HasAny() {
		return r, nil
	}
	if mV != nil {
		m := mV.Interface().(C)
		r.Data = append(r.Data, m)
		r.ValidateStruct(-1, m)
	}

	return r, nil
}
