package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"golang.org/x/net/html/charset"
)

//DetermineEncodingFromReader tries to get the encoding from a reader
func DetermineEncodingFromReader(file multipart.File, size int64) string {
	var mLen int64 = 1024
	if size < mLen {
		mLen = size
	}

	b, err := bufio.NewReader(file).Peek(int(mLen))
	if err != nil {
		return ""
	}

	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("File-Seek1:", err)
		return ""
	}

	_, name, _ := charset.DetermineEncoding(b, "")
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("File-Seek2:", err)
		return ""
	}
	return name
}

func GetFileContentType(out *os.File) (string, error) {
	if out == nil {
		return "", errors.New("No file specified")
	}
	var mLen int64 = 512 // Only the first 512 bytes are used to sniff the content type.

	fs, err := out.Stat()
	if err != nil {
		return "", err
	}
	if fs.Size() < mLen {
		mLen = fs.Size()
	}

	buffer := make([]byte, mLen)

	if _, err := out.Seek(0, 0); err != nil {
		return "", err
	}

	_, err = out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	out.Seek(0, 0) //no matter of error we try to go to the file beginning

	return contentType, nil
}

//CreateDirIfNotExists creates an directory if it does not exists
func CreateDirIfNotExists(path string) error {
	if !ExistDir(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf(fmt.Sprintf("Could not create dir <%s>\n", path))
			return err
		}
	}
	return nil
}

//ExistDir checks if an directory exists
func ExistDir(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

//FileExists checks if file exists
func FileExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	}
	return false
}
