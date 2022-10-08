package uploader

import "fmt"

type ErrUploadMessage struct {
	StrMsg string
}

func UploadErrMessageForFile(message string, fileName string) *ErrUploadMessage {
	return &ErrUploadMessage{
		StrMsg: fmt.Sprintf(message, fileName),
	}
}

func UploadErrMessage(message string) *ErrUploadMessage {
	return &ErrUploadMessage{
		StrMsg: message,
	}
}

func (r *ErrUploadMessage) Error() string {
	return r.StrMsg
}
