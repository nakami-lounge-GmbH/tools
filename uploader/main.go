package uploader

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

	mtype := http.DetectContentType(buff)

	if !typeChecker.IsAllowed(mtype) {
		return nil, "", UploadErrMessage(fmt.Sprintf("Datei %s entspricht nicht dem geforderten Format: %s", f.Filename, typeChecker.GetTypesString()))
	}

	return buff, f.Filename, nil
}
