package uploader

import (
	"encoding/binary"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"mime/multipart"
)

func GetUploadFile(f *multipart.FileHeader, typeChecker AllowedFileTypes, maxMBSize int64) ([]byte, string, error) {
	/*f, err := c.FormFile(fileParam)
	if err != nil {
		if err != http.ErrMissingFile {
			return nil, "", fmt.Errorf("reading file %s, %w", fileParam, err)
		} else {
			return nil, "", nil
		}
	}*/

	if f.Size > maxMBSize*1024*1024 {
		return nil, "", UploadErrMessage(fmt.Sprintf("Datei '%s' ist zu gross. Max %dMB erlaubt", f.Filename, maxMBSize))
	}

	ff, err := f.Open()
	if err != nil {
		return nil, "", fmt.Errorf("opening %s %w", f.Filename, err)
	}
	defer ff.Close()

	buff, err := io.ReadAll(ff)
	if err != nil {
		return nil, "", fmt.Errorf("reading buffer for file %s %w", f.Filename, err)
	}

	//mtype := http.DetectContentType(buff)
	mtype := mimetype.Detect(buff)

	if !typeChecker.IsAllowed(mtype.String()) {
		return nil, "", UploadErrMessage(fmt.Sprintf("Datei %s entspricht nicht dem geforderten Format: %s", f.Filename, typeChecker.GetTypesString()))
	}

	return buff, f.Filename, nil
}

func CheckFileTypeAndSize(data []byte, filename string, typeChecker AllowedFileTypes, maxMBSize int) error {
	size := binary.Size(data)
	if float64(size)/float64(1024)/float64(1024) > float64(maxMBSize) {
		return UploadErrMessage(fmt.Sprintf("Datei '%s' ist zu gross. Max %dMB erlaubt", filename, maxMBSize))
	}

	//mtype := http.DetectContentType(data)
	mtype := mimetype.Detect(data)

	if !typeChecker.IsAllowed(mtype.String()) {
		return UploadErrMessage(fmt.Sprintf("Datei %s entspricht nicht dem geforderten Format: %s", filename, typeChecker.GetTypesString()))
	}

	return nil
}
