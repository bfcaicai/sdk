package utils

import (
	"crypto/md5"
	"fmt"
	"io"
)

func Md5String(md5Str string) string {
	md5 := md5.New()
	io.WriteString(md5, md5Str)
	return fmt.Sprintf("%x", md5.Sum(nil))
}
