package helpers

import "math/rand"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

//RandomString generates a random string of len count
func RandomString(count int) string {
	b := make([]byte, count)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

//StringMax returns the maximam number of strings
func StringMax(s string, l int) string {
	if len(s) < l {
		l = len(s)
	}
	return s[:l]
}
