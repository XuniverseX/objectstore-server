package util

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func FileMd5(file *os.File) string {
	file.Seek(0, 0)
	m := md5.New()
	io.Copy(m, file)
	return hex.EncodeToString(m.Sum(nil))
}
